package controllers

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func setCondition(
	conditions *[]metav1.Condition,
	condType string,
	status metav1.ConditionStatus,
	reason, message string,
) {
	now := metav1.Now()

	for i, c := range *conditions {
		if c.Type == condType {
			(*conditions)[i] = metav1.Condition{
				Type:               condType,
				Status:             status,
				Reason:             reason,
				Message:            message,
				LastTransitionTime: now,
			}
			return
		}
	}

	*conditions = append(*conditions, metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	})
}
