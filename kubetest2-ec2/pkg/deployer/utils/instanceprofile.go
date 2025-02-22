package utils

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"k8s.io/klog/v2"
)

func GetInstanceProfileArn(svc *iam.IAM, instanceProfileName string) (string, error) {
	listInstanceProfilesInput := &iam.ListInstanceProfilesInput{
		PathPrefix: aws.String("/kubetest2/"),
	}
	listInstanceProfilesResult, err := svc.ListInstanceProfiles(listInstanceProfilesInput)
	if err != nil {
		return "", err
	}
	if len(listInstanceProfilesResult.InstanceProfiles) > 0 {
		for _, profile := range listInstanceProfilesResult.InstanceProfiles {
			if *profile.InstanceProfileName == instanceProfileName {
				return *profile.Arn, nil
			}
		}
	}
	return "", fmt.Errorf("unable to find Arn for %s instance profile", instanceProfileName)
}

func EnsureInstanceProfile(svc *iam.IAM, instanceProfileName string, roleName string) error {

	listInstanceProfilesInput := &iam.ListInstanceProfilesInput{
		PathPrefix: aws.String("/kubetest2/"),
	}

	listInstanceProfilesResult, err := svc.ListInstanceProfiles(listInstanceProfilesInput)
	if err != nil {
		return err
	}
	if len(listInstanceProfilesResult.InstanceProfiles) > 0 {
		for _, profile := range listInstanceProfilesResult.InstanceProfiles {
			if *profile.InstanceProfileName == instanceProfileName {
				klog.Infof("%s instance profile exists already ARN: %s\n", instanceProfileName, *profile.Arn)
				return nil
			}
		}
	} else {
		klog.Infof("did not find any pre-existing %s. creating %s...\n", instanceProfileName, instanceProfileName)
	}

	createInput := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
		Path:                aws.String("/kubetest2/"),
	}

	createResult, err := svc.CreateInstanceProfile(createInput)
	if err != nil {
		return fmt.Errorf("unable to create instance profile : %w", err)
	}
	klog.Infof("created instance profile: %v\n", *createResult.InstanceProfile.Arn)

	listProfilesForRoleInput := &iam.ListInstanceProfilesForRoleInput{RoleName: aws.String(roleName)}
	listProfilesForRoleResult, err := svc.ListInstanceProfilesForRole(listProfilesForRoleInput)
	if err != nil {
		return fmt.Errorf("unable to list instance profile for role: %w", err)
	}
	if len(listProfilesForRoleResult.InstanceProfiles) > 0 {
		klog.Infof("found instance profile %s for role %s already", instanceProfileName, roleName)
		return nil
	}

	addInput := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
		RoleName:            aws.String(roleName),
	}
	_, err = svc.AddRoleToInstanceProfile(addInput)
	if err != nil {
		return fmt.Errorf("unable to add role to instance profile : %w", err)
	}
	klog.Infof("added role %s to instance profile %s\n", roleName, instanceProfileName)
	return nil
}
