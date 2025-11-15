package main

import (
	"strconv"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func buildStartKeyboard(api *maxbot.Api, userId int64) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	if isCompany(userId) {
		return buildCompanyStartKeyboard(api, keyboard)
	} else if isStudent(userId) {
		return buildStudentStartKeyboard(api, keyboard)
	} else if isUni(userId) {
		return buildUniStartKeyboard(api, keyboard)
	} else {
		return buildDefaultStartKeyboard(api, keyboard)
	}
}

func buildDefaultStartKeyboard(api *maxbot.Api, kb *maxbot.Keyboard) *maxbot.Keyboard {
	kb.
		AddRow().
		AddCallback("Компания", schemes.POSITIVE, "role_company").
		AddCallback("Университет", schemes.POSITIVE, "role_uni").
		AddCallback("Студент", schemes.POSITIVE, "role_student")
	return kb
}

func buildCompanyStartKeyboard(api *maxbot.Api, kb *maxbot.Keyboard) *maxbot.Keyboard {
	kb.
		AddRow().
		AddCallback("Добавить тему для ВКР", schemes.POSITIVE, "add_topic").
		AddCallback("Добавить практику", schemes.POSITIVE, "add_intern")
	kb.
		AddRow().
		AddCallback("Посмотреть заявки", schemes.POSITIVE, "view_pending_company")
	return kb
}

func buildUniStartKeyboard(api *maxbot.Api, kb *maxbot.Keyboard) *maxbot.Keyboard {
	kb.
		AddRow().
		AddCallback("Согласовать темы для ВКР", schemes.POSITIVE, "accept_topics").
		AddCallback("Согласовать практики", schemes.POSITIVE, "accept_interns")
	kb.
		AddRow().
		AddCallback("Посмотреть заявки в обработке", schemes.POSITIVE, "view_pending_uni")
	return kb
}

func buildStudentStartKeyboard(api *maxbot.Api, kb *maxbot.Keyboard) *maxbot.Keyboard {
	kb.
		AddRow().
		AddCallback("Подать заявку на ВКР", schemes.POSITIVE, "apply_diploma").
		AddCallback("Подать заявку на практику", schemes.POSITIVE, "apply_intern")
	return kb
}

func buildTargetKeyboard(api *maxbot.Api, prefix string) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback("General pool", schemes.POSITIVE, prefix+"general")
	for name := range unis {
		keyboard.AddRow().AddCallback(name, schemes.POSITIVE, prefix+name)
	}
	return keyboard
}

func buildUniListKeyboard(api *maxbot.Api, prefix string) *maxbot.Keyboard {
	if len(unis) == 0 {
		return nil
	}
	keyboard := api.Messages.NewKeyboardBuilder()
	for name := range unis {
		keyboard.AddRow().AddCallback(name, schemes.POSITIVE, prefix+name)
	}
	return keyboard
}

func buildAcceptTopicsKeyboard(api *maxbot.Api, uniName string) *maxbot.Keyboard {
	var hasTopics bool
	keyboard := api.Messages.NewKeyboardBuilder()

	// General
	for _, topic := range diplomaTopics {
		if topic.Status == "available" && !contains(uniDiplomaPools[uniName], topic.ID) {
			keyboard.AddRow().AddCallback(topic.Name+" (ID: "+strconv.Itoa(topic.ID)+")", schemes.POSITIVE, "accept_topic_"+strconv.Itoa(topic.ID))
			hasTopics = true
		}
	}
	// Specific
	for _, topic := range specificDiplomaTopics[uniName] {
		if topic.Status == "available" && !contains(uniDiplomaPools[uniName], topic.ID) {
			keyboard.AddRow().AddCallback(topic.Name+" (ID: "+strconv.Itoa(topic.ID)+")", schemes.POSITIVE, "accept_topic_"+strconv.Itoa(topic.ID))
			hasTopics = true
		}
	}
	if !hasTopics {
		return nil
	}
	return keyboard
}

func buildAcceptInternsKeyboard(api *maxbot.Api, uniName string) *maxbot.Keyboard {
	var hasInterns bool
	keyboard := api.Messages.NewKeyboardBuilder()

	// General
	for _, intern := range internshipOffers {
		if intern.Status == "available" && intern.Places > 0 && !contains(uniInternshipPools[uniName], intern.ID) {
			keyboard.AddRow().AddCallback(intern.Name+" (ID: "+strconv.Itoa(intern.ID)+")", schemes.POSITIVE, "accept_intern_"+strconv.Itoa(intern.ID))
			hasInterns = true
		}
	}
	// Specific
	for _, intern := range specificInternships[uniName] {
		if intern.Status == "available" && intern.Places > 0 && !contains(uniInternshipPools[uniName], intern.ID) {
			keyboard.AddRow().AddCallback(intern.Name+" (ID: "+strconv.Itoa(intern.ID)+")", schemes.POSITIVE, "accept_intern_"+strconv.Itoa(intern.ID))
			hasInterns = true
		}
	}
	if !hasInterns {
		return nil
	}
	return keyboard
}

func buildApplyDiplomaKeyboard(api *maxbot.Api, uniName string) *maxbot.Keyboard {
	var hasTopics bool
	keyboard := api.Messages.NewKeyboardBuilder()
	for _, tid := range uniDiplomaPools[uniName] {
		topic := findDiplomaTopicByID(tid)
		if topic != nil && topic.Status == "available" {
			keyboard.AddRow().AddCallback(topic.Name+" (ID: "+strconv.Itoa(topic.ID)+")", schemes.POSITIVE, "apply_diploma_"+strconv.Itoa(topic.ID))
			hasTopics = true
		}
	}
	if !hasTopics {
		return nil
	}
	return keyboard
}

func buildApplyInternKeyboard(api *maxbot.Api, uniName string) *maxbot.Keyboard {
	var hasInterns bool
	keyboard := api.Messages.NewKeyboardBuilder()
	for _, iid := range uniInternshipPools[uniName] {
		intern := findInternByID(iid)
		if intern != nil && intern.Status == "available" && intern.Places > 0 {
			keyboard.AddRow().AddCallback(intern.Name+" (ID: "+strconv.Itoa(intern.ID)+")", schemes.POSITIVE, "apply_intern_"+strconv.Itoa(intern.ID))
			hasInterns = true
		}
	}
	if !hasInterns {
		return nil
	}
	return keyboard
}
