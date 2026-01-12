package utils

import "fmt"

func GetApplicationKey(contestId, userId int64) string {
	return fmt.Sprintf("contest:%d:application:%d", contestId, userId)
}

func GetPendingKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:applications:pending", contestId)
}

func GetAcceptedKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:applications:accepted", contestId)
}

func GetRejectedKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:applications:rejected", contestId)
}

func GetUserApplicationsKey(userId int64) string {
	return fmt.Sprintf("user:%d:applications", userId)
}

func GetContestPatternKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:application*", contestId)
}
