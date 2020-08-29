package controllers

import (
	"context"
	"fmt"
	"strings"
	tsunamiV1 "tsunami-controller/api/v1"

	"github.com/go-logr/logr"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	ownerKey            = ".metadata.controller"
	defaultTsunamiImage = "docker.pkg.github.com/kaidotorg/workspace/tsunami:v1.3.0"
)

type ScanReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	Recorder     record.EventRecorder
	TsunamiImage string
}

func (r *ScanReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	scan := &tsunamiV1.Scan{}
	ctx := context.Background()
	logger := r.Log.WithValues("scan", req.NamespacedName)
	if err := r.Client.Get(ctx, req.NamespacedName, scan); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.cleanupOwnedResources(ctx, scan); err != nil {
		return ctrl.Result{}, err
	}

	f := func(podIP string) error {
		var job batchV1.Job
		if err := r.Client.Get(
			ctx,
			client.ObjectKey{
				Name:      req.Name + "-scan-" + podIP,
				Namespace: req.Namespace,
			},
			&job,
		); errors.IsNotFound(err) {
			job = *r.buildJob(scan, podIP)
			if err := controllerutil.SetControllerReference(scan, &job, r.Scheme); err != nil {
				return err
			}
			if err := r.Create(ctx, &job); err != nil {
				return err
			}
			r.Recorder.Eventf(scan, coreV1.EventTypeNormal, "SuccessfulCreated", "Created job: %q", job.Name)
			logger.V(1).Info("create", "job", job)
		} else if err != nil {
			return err
		}
		return nil
	}

	matchingLabels := client.MatchingLabels{}
	for k, v := range scan.Spec.MatchLabels {
		matchingLabels[k] = v
	}

	var pods v1.PodList
	if err := r.Client.List(
		ctx,
		&pods,
		client.InNamespace(scan.Namespace),
		matchingLabels,
	); err != nil {
		return ctrl.Result{}, err
	}
	for _, pod := range pods.Items {
		podIP := pod.Status.PodIP
		if podIP == "" {
			continue
		}
		err := f(podIP)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ScanReconciler) buildJob(scan *tsunamiV1.Scan, podIP string) *batchV1.Job {
	appLabel := scan.Name + "-scan"

	labels := map[string]string{
		"app": appLabel,
	}
	for k, v := range scan.Spec.Template.ObjectMeta.Labels {
		labels[k] = v
	}
	scan.Spec.Template.ObjectMeta.Labels = labels

	var tsunamiImage string
	if r.TsunamiImage == "" {
		tsunamiImage = defaultTsunamiImage
	} else {
		tsunamiImage = r.TsunamiImage
	}

	return &batchV1.Job{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      scan.Name + "-scan-" + podIP,
			Namespace: scan.Namespace,
		},
		Spec: batchV1.JobSpec{
			Parallelism: func(i int32) *int32 {
				return &i
			}(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: scan.Spec.Template.ObjectMeta,
				Spec: v1.PodSpec{
					TopologySpreadConstraints: []v1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: v1.ScheduleAnyway,
							LabelSelector: &metaV1.LabelSelector{
								MatchLabels: map[string]string{
									"app": appLabel,
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "scanner",
							Image: tsunamiImage,
							Args: []string{fmt.Sprintf(
								"--ip-v4-target=%s",
								podIP,
							)},
							ImagePullPolicy: v1.PullIfNotPresent,
							Resources:       scan.Spec.ScannerContainerSpec.Resources,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
}

func (r *ScanReconciler) cleanupOwnedResources(ctx context.Context, scan *tsunamiV1.Scan) error {
	var jobs batchV1.JobList
	if err := r.Client.List(
		ctx,
		&jobs,
		client.InNamespace(scan.Namespace),
		client.MatchingFields{ownerKey: scan.Name},
	); err != nil {
		return err
	}

	for _, job := range jobs.Items {
		job := job

		if strings.HasPrefix(job.Name, scan.Name+"-scan-") {
			continue
		}

		if err := r.Client.Delete(ctx, &job); err != nil {
			return err
		}
		r.Recorder.Eventf(scan, coreV1.EventTypeNormal, "SuccessfulDeleted", "Deleted job: %q", job.Name)
	}

	return nil
}

func (r *ScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchV1.Job{}, ownerKey, func(rawObj runtime.Object) []string {
		job := rawObj.(*batchV1.Job)
		owner := metaV1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		if owner.Kind != "Scan" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&tsunamiV1.Scan{}).
		Owns(&batchV1.Job{}).
		Complete(r)
}
