package main

import (
	"context"
	"strconv"
	"strings"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func notifyCompanyForApp(api *maxbot.Api, ctx context.Context, app *Application) {
	companyUser := findUserByRole("company:" + app.CompanyINN)
	if companyUser == 0 {
		return
	}
	itemName := ""
	prefix := "company_accept_diploma_"
	if app.Type == "intern" {
		intern := findInternByID(app.ItemID)
		if intern != nil {
			itemName = intern.Name
		}
		prefix = "company_accept_intern_"
	} else {
		topic := findDiplomaTopicByID(app.ItemID)
		if topic != nil {
			itemName = topic.Name
		}
	}
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Accept", schemes.POSITIVE, prefix+strconv.Itoa(app.ID)).
		AddCallback("Reject", schemes.NEGATIVE, "company_reject_"+strconv.Itoa(app.ID))
	student := students[app.StudentID]
	msg := "Студент " + student.FIO + " подал заявка на " + app.Type + " : " + itemName
	sendMessageWithKeyboard(api, ctx, companyUser, msg, keyboard)
}

func notifyUniForApp(api *maxbot.Api, ctx context.Context, app *Application) {
	uniUser := findUserByRole("uni:" + app.UniName)
	if uniUser == 0 {
		return
	}
	itemName := ""
	prefix := "uni_approve_diploma_"
	if app.Type == "intern" {
		intern := findInternByID(app.ItemID)
		if intern != nil {
			itemName = intern.Name
		}
		prefix = "uni_approve_intern_"
	} else {
		topic := findDiplomaTopicByID(app.ItemID)
		if topic != nil {
			itemName = topic.Name
		}
	}
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Approve", schemes.POSITIVE, prefix+strconv.Itoa(app.ID)).
		AddCallback("Reject", schemes.NEGATIVE, "uni_reject_"+strconv.Itoa(app.ID))
	student := students[app.StudentID]
	msg := "Студент " + student.FIO + " подал заявку на " + app.Type + " : " + itemName + ", Компания согласовала."
	sendMessageWithKeyboard(api, ctx, uniUser, msg, keyboard)
}

func notifyUni(api *maxbot.Api, ctx context.Context, uniName string, message string) {
	uniUser := findUserByRole("uni:" + uniName)
	if uniUser != 0 {
		sendMessage(api, ctx, uniUser, message)
	}
}

func findDiplomaTopicByID(id int) *DiplomaTopic {
	for i := range diplomaTopics {
		if diplomaTopics[i].ID == id {
			return &diplomaTopics[i]
		}
	}
	for _, topics := range specificDiplomaTopics {
		for i := range topics {
			if topics[i].ID == id {
				return &topics[i]
			}
		}
	}
	return nil
}

func findInternByID(id int) *Internship {
	for i := range internshipOffers {
		if internshipOffers[i].ID == id {
			return &internshipOffers[i]
		}
	}
	for _, interns := range specificInternships {
		for i := range interns {
			if interns[i].ID == id {
				return &interns[i]
			}
		}
	}
	return nil
}

func findAppByID(id int) *Application {
	for i := range applications {
		if applications[i].ID == id {
			return &applications[i]
		}
	}
	return nil
}

func findUserByRole(role string) int64 {
	for u, r := range userRoles {
		if r == role {
			return u
		}
	}
	return 0
}

func isCompany(userId int64) bool {
	role, ok := userRoles[userId]
	return ok && strings.HasPrefix(role, "company:")
}

func getCompanyINN(userId int64) string {
	role, ok := userRoles[userId]
	if !ok || !strings.HasPrefix(role, "company:") {
		return ""
	}
	return strings.TrimPrefix(role, "company:")
}

func isStudent(userId int64) bool {
	role, ok := userRoles[userId]
	return ok && role == "student"
}

func isUni(userId int64) bool {
	role, ok := userRoles[userId]
	return ok && strings.HasPrefix(role, "uni:")
}

func getUniName(userId int64) string {
	role, ok := userRoles[userId]
	if !ok || !strings.HasPrefix(role, "uni:") {
		return ""
	}
	return strings.TrimPrefix(role, "uni:")
}

func contains(list []int, val int) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func ensureTempData(userId int64) {
	if userTempData[userId] == nil {
		userTempData[userId] = make(map[string]string)
	}
}

func cleanupUserData(userId int64) {
	delete(userStates, userId)
	delete(userTempData, userId)
}

func sendMessage(api *maxbot.Api, ctx context.Context, userId int64, text string) {
	_, _ = api.Messages.Send(ctx, maxbot.NewMessage().SetUser(userId).SetText(text))
}

func sendMessageWithKeyboard(api *maxbot.Api, ctx context.Context, userId int64, text string, keyboard *maxbot.Keyboard) {
	msg := maxbot.NewMessage().SetUser(userId).SetText(text)
	if keyboard != nil {
		msg.AddKeyboard(keyboard)
	}
	api.Messages.Send(ctx, msg)
}
