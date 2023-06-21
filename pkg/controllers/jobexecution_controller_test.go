package controllers

import (
	"context"
	"fmt"
	"time"

	dispatcherv1alpha1 "github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("JobExecution controller", func() {
	iteration := 0

	var (
		namespace         *corev1.Namespace
		typeNamespaceName types.NamespacedName
		jobTemplate       *dispatcherv1alpha1.JobTemplate
		namespaceName     string
	)

	const (
		jobExecutionName = "test-jobexecution"
		jobTemplateName  = "test-jobtemplate"
		namespacePrefix  = "dispatcher-tests-"
	)

	ctx := context.Background()

	BeforeEach(func() {
		namespaceName = fmt.Sprintf("%s%d", namespacePrefix, iteration)
		iteration++

		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespaceName,
				Namespace: namespaceName,
			},
		}

		typeNamespaceName = types.NamespacedName{
			Name:      jobExecutionName,
			Namespace: namespaceName,
		}

		jobTemplate = &dispatcherv1alpha1.JobTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobTemplateName,
				Namespace: namespaceName,
			},
			Spec: dispatcherv1alpha1.JobTemplateSpec{
				JobTemplateSpec: batchv1.JobTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: jobExecutionName,
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name:    "pi",
										Image:   "busybox",
										Command: []string{"exit", "0"},
										Env: []corev1.EnvVar{
											corev1.EnvVar{
												Name:  "PAYLOAD",
												Value: `{{ if fromJson .Payload }}{{ .Payload | fromJson | toJson }}{{ else }}{{ .Payload }}{{ end }}`,
											},
										},
									},
								},
								RestartPolicy: corev1.RestartPolicyNever,
							},
						},
					},
				},
			},
		}

		err := k8sClient.Create(ctx, namespace)
		Expect(err).To(Not(HaveOccurred()))

		err = k8sClient.Create(ctx, jobTemplate)
		Expect(err).To(Not(HaveOccurred()))
	})

	AfterEach(func() {
		_ = k8sClient.Delete(ctx, namespace)
		_ = k8sClient.Delete(ctx, jobTemplate)
	})

	It("reconciles a custom resource for JobTemplate", func() {
		By("Creating the JobExecution")
		jobExecution := &dispatcherv1alpha1.JobExecution{
			Spec: dispatcherv1alpha1.JobExecutionSpec{
				JobTemplateName: jobTemplateName,
			},
		}
		err := k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		if err != nil && errors.IsNotFound(err) {
			jobExecution := &dispatcherv1alpha1.JobExecution{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobExecutionName,
					Namespace: namespace.Name,
				},
				Spec: dispatcherv1alpha1.JobExecutionSpec{
					JobTemplateName: jobTemplateName,
					Payload:         "test",
				},
			}

			err = k8sClient.Create(ctx, jobExecution)
			Expect(err).To(Not(HaveOccurred()))
		}

		By("Checking if the custom resource was created")
		Eventually(func() error {
			found := &dispatcherv1alpha1.JobExecution{}
			return k8sClient.Get(ctx, typeNamespaceName, found)
		}, time.Minute, time.Second).Should(Succeed())

		By("Running the reconciliation")
		jobExecutionReconciler := &JobExecutionReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
		res, err := jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(Not(HaveOccurred()))
		Expect(res.Requeue).To(BeTrue())

		By("Checking if the status of the JobExecution is waiting")
		k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		Expect(jobExecution.Status.Phase).To(Equal(dispatcherv1alpha1.JobExecutionWaitingPhase))

		job := &batchv1.Job{}
		By("Checking if the Job from the JobExecution was created")
		Eventually(func() error {
			return k8sClient.Get(ctx, typeNamespaceName, job)
		}, time.Minute, time.Second).Should(Succeed())

		By("Checking the labels from the generated Job")
		Expect(job.ObjectMeta.Labels).To(HaveKeyWithValue("controller-uid", string(jobExecution.UID)))
		Expect(job.ObjectMeta.Labels).To(HaveKeyWithValue("job-execution-name", jobExecutionName))

		By("Checking the reference to the Job")
		Expect(jobExecution.Status.Job.Kind).To(Equal("Job"))
		Expect(jobExecution.Status.Job.Name).To(Equal(job.Name))
		Expect(jobExecution.Status.Job.Namespace).To(Equal(job.Namespace))
		Expect(jobExecution.Status.Job.UID).To(Equal(job.UID))

		By("Updating the JobExecution status when Job is running")
		now := metav1.Now()
		job.Status.StartTime = &now
		k8sClient.Status().Update(ctx, job)
		res, err = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(Not(HaveOccurred()))
		Expect(res.RequeueAfter).To(Not(BeNil()))
		k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		Expect(jobExecution.Status.Phase).To(Equal(dispatcherv1alpha1.JobExecutionActivePhase))

		By("Updating the JobExecution status when Job finished running")
		job.Status.CompletionTime = &now
		k8sClient.Status().Update(ctx, job)
		res, err = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(Not(HaveOccurred()))
		Expect(res.RequeueAfter).To(Not(BeNil()))
		k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		Expect(jobExecution.Status.Phase).To(Equal(dispatcherv1alpha1.JobExecutionCompletedPhase))

		By("Deleting the JobExecution once the Job is removed")
		job.Labels["controller-uid"] = ""
		k8sClient.Update(ctx, job)
		_, err = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(Not(HaveOccurred()))
		err = k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("sets the job name when GenerateName is set", func() {
		By("Setting up the JobTemplate")
		jobTemplate.Spec.JobTemplateSpec.ObjectMeta.Name = ""
		jobTemplate.Spec.JobTemplateSpec.ObjectMeta.GenerateName = "test-"
		k8sClient.Update(ctx, jobTemplate)

		By("Setting up the JobExecution")
		jobExecution := &dispatcherv1alpha1.JobExecution{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobExecutionName,
				Namespace: namespace.Name,
			},
			Spec: dispatcherv1alpha1.JobExecutionSpec{
				JobTemplateName: jobTemplateName,
				Payload:         "test",
			},
		}

		k8sClient.Create(ctx, jobExecution)

		By("Running the reconciliation")
		jobExecutionReconciler := &JobExecutionReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
		_, _ = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		k8sClient.Get(ctx, typeNamespaceName, jobExecution)

		typeNamespaceName.Name = jobExecution.Status.Job.Name
		job := &batchv1.Job{}
		By("Checking if the Job from the JobExecution was created")
		Eventually(func() error {
			return k8sClient.Get(ctx, typeNamespaceName, job)
		}, time.Minute, time.Second).Should(Succeed())

		By("Checking the generated name for the Job")
		Expect(job.Name).To(HavePrefix("test-"))
	})

	It("generates a new if neither Name nor GenerateName are set", func() {
		By("Setting up the JobTemplate")
		jobTemplate.Spec.JobTemplateSpec.ObjectMeta.Name = ""
		jobTemplate.Spec.JobTemplateSpec.ObjectMeta.GenerateName = ""
		k8sClient.Update(ctx, jobTemplate)

		By("Setting up the JobExecution")
		jobExecution := &dispatcherv1alpha1.JobExecution{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobExecutionName,
				Namespace: namespace.Name,
			},
			Spec: dispatcherv1alpha1.JobExecutionSpec{
				JobTemplateName: jobTemplateName,
				Payload:         "test",
			},
		}
		k8sClient.Create(ctx, jobExecution)

		By("Running the reconciliation")
		jobExecutionReconciler := &JobExecutionReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
		_, _ = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		k8sClient.Get(ctx, typeNamespaceName, jobExecution)

		typeNamespaceName.Name = jobExecution.Status.Job.Name
		job := &batchv1.Job{}
		By("Checking if the Job from the JobExecution was created")
		Eventually(func() error {
			return k8sClient.Get(ctx, typeNamespaceName, job)
		}, time.Minute, time.Second).Should(Succeed())

		By("Checking the generated name for the Job")
		Expect(job.Name).To(HavePrefix(jobExecutionName + "-"))
	})

	It("sets its state as fail if Job fails to run", func() {
		By("Creating the JobExecution")
		jobExecution := &dispatcherv1alpha1.JobExecution{
			Spec: dispatcherv1alpha1.JobExecutionSpec{
				JobTemplateName: jobTemplateName,
			},
		}
		err := k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		if err != nil && errors.IsNotFound(err) {
			jobExecution := &dispatcherv1alpha1.JobExecution{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobExecutionName,
					Namespace: namespace.Name,
				},
				Spec: dispatcherv1alpha1.JobExecutionSpec{
					JobTemplateName: jobTemplateName,
					Payload:         "test",
				},
			}

			err = k8sClient.Create(ctx, jobExecution)
			Expect(err).To(Not(HaveOccurred()))
		}

		By("Checking if the custom resource was created")
		Eventually(func() error {
			found := &dispatcherv1alpha1.JobExecution{}
			return k8sClient.Get(ctx, typeNamespaceName, found)
		}, time.Minute, time.Second).Should(Succeed())

		By("Running the reconciliation")
		jobExecutionReconciler := &JobExecutionReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
		res, err := jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(Not(HaveOccurred()))
		Expect(res.Requeue).To(BeTrue())

		job := &batchv1.Job{}
		By("Checking if the Job from the JobExecution was created")
		Eventually(func() error {
			return k8sClient.Get(ctx, typeNamespaceName, job)
		}, time.Minute, time.Second).Should(Succeed())

		By("Updating the JobExecution status when Job is running")
		now := metav1.Now()
		job.Status.StartTime = &now
		k8sClient.Status().Update(ctx, job)
		res, err = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(Not(HaveOccurred()))
		Expect(res.RequeueAfter).To(Not(BeNil()))
		k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		Expect(jobExecution.Status.Phase).To(Equal(dispatcherv1alpha1.JobExecutionActivePhase))

		By("Updating the JobExecution status when Job finished running")
		job.Status.Conditions = []batchv1.JobCondition{{
			Type:   batchv1.JobFailed,
			Status: corev1.ConditionTrue,
		}}
		k8sClient.Status().Update(ctx, job)
		_, _ = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		Expect(jobExecution.Status.Phase).To(Equal(dispatcherv1alpha1.JobExecutionFailedPhase))

		By("Deleting the JobExecution once the Job is removed")
		job.Labels["controller-uid"] = ""
		k8sClient.Update(ctx, job)
		_, err = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(Not(HaveOccurred()))
		err = k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("fails if no jobTemplate is found", func() {
		By("Creating the JobExecution")
		jobExecution := &dispatcherv1alpha1.JobExecution{
			Spec: dispatcherv1alpha1.JobExecutionSpec{
				JobTemplateName: "not-found",
			},
		}
		err := k8sClient.Get(ctx, typeNamespaceName, jobExecution)
		if err != nil && errors.IsNotFound(err) {
			jobExecution := &dispatcherv1alpha1.JobExecution{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobExecutionName,
					Namespace: namespace.Name,
				},
				Spec: dispatcherv1alpha1.JobExecutionSpec{
					JobTemplateName: "not-found",
					Payload:         "test",
				},
			}

			err = k8sClient.Create(ctx, jobExecution)
			Expect(err).To(Not(HaveOccurred()))
		}

		By("Checking if the custom resource was created")
		Eventually(func() error {
			found := &dispatcherv1alpha1.JobExecution{}
			return k8sClient.Get(ctx, typeNamespaceName, found)
		}, time.Minute, time.Second).Should(Succeed())

		By("Running the reconciliation")
		jobExecutionReconciler := &JobExecutionReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
		_, err = jobExecutionReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typeNamespaceName,
		})
		Expect(err).To(HaveOccurred())
	})
})
