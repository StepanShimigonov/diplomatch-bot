package main

import (
	"context"
	"strconv"
	"strings"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func handleMessageCreated(api *maxbot.Api, ctx context.Context, upd *schemes.MessageCreatedUpdate) {
	userId := upd.Message.Sender.UserId
	command := upd.GetCommand()
	state, hasState := userStates[userId]

	if hasState {
		handleState(api, ctx, userId, state, upd)
	} else if command != "" {
		handleCommand(api, ctx, userId, command, upd)
	} else {
		sendMessage(api, ctx, userId, "Введите /start")
	}
}

func handleCommand(api *maxbot.Api, ctx context.Context, userId int64, command string, upd *schemes.MessageCreatedUpdate) {
	switch command {
	case "/start":
		role := userRoles[userId]
		msg := "Добро пожаловать!"
		if role == "" {
			msg = "Выберите тип учетной записи:"
		} else if strings.HasPrefix(role, "Компания:") {
			msg = "Меню Компании:"
		} else if strings.HasPrefix(role, "Университет:") {
			msg = "Меню Университета:"
		} else if role == "Студент" {
			msg = "Меню Студента:"
		}
		kb := buildStartKeyboard(api, userId)
		sendMessageWithKeyboard(api, ctx, userId, msg, kb)
	default:
		sendMessage(api, ctx, userId, "Неизвестная команда.")
	}
}

func handleState(api *maxbot.Api, ctx context.Context, userId int64, state string, upd *schemes.MessageCreatedUpdate) {
	text := upd.Message.Body.Text
	ensureTempData(userId)

	switch state {
	case "reg_company_name":
		userTempData[userId]["name"] = text
		userStates[userId] = "reg_company_inn"
		sendMessage(api, ctx, userId, "Введите ИНН:")
	case "reg_company_inn":
		inn := text
		if _, exists := companies[inn]; exists {
			sendMessage(api, ctx, userId, "Компания с таким ИНН уже зарегестрирована. Свяжитесь с Администратором, если это ошибка, либо попробуйте снова.")
			return
		}
		companies[inn] = Company{Name: userTempData[userId]["name"], INN: inn}
		userRoles[userId] = "company:" + inn
		cleanupUserData(userId)
		sendMessage(api, ctx, userId, "Регистрация прошла успешно.")
		kb := buildStartKeyboard(api, userId)
		sendMessageWithKeyboard(api, ctx, userId, "Меню Компании:", kb)
	case "reg_uni_name":
		name := text
		if _, exists := unis[name]; exists {
			sendMessage(api, ctx, userId, "Данные Университет уже зарегестрирован. Свяжитесь с Администратором, если это ошибка, либо попробуйте снова..")
			return
		}
		unis[name] = Uni{Name: name}
		userRoles[userId] = "uni:" + name
		cleanupUserData(userId)
		sendMessage(api, ctx, userId, "Университет успешно зарегестрирован.")
		kb := buildStartKeyboard(api, userId)
		sendMessageWithKeyboard(api, ctx, userId, "Меню Университета:", kb)
	case "student_course":
		userTempData[userId]["course"] = text
		userStates[userId] = "student_fio"
		sendMessage(api, ctx, userId, "Введите свое ФИО:")
	case "student_fio":
		fio := text
		students[userId] = Student{
			UniName: userTempData[userId]["uni"],
			Course:  userTempData[userId]["course"],
			FIO:     fio,
		}
		userRoles[userId] = "student"
		cleanupUserData(userId)
		sendMessage(api, ctx, userId, "Вы успешно зарегестрированы.")
		kb := buildStartKeyboard(api, userId)
		sendMessageWithKeyboard(api, ctx, userId, "Меню Студента:", kb)
	case "add_topic_name":
		userTempData[userId]["topic_name"] = text
		userStates[userId] = "add_topic_desc"
		sendMessage(api, ctx, userId, "Введите описание темы для ВКР:")
	case "add_topic_desc":
		userTempData[userId]["topic_desc"] = text
		userStates[userId] = "add_topic_target"
		keyboard := buildTargetKeyboard(api, "target_")
		sendMessageWithKeyboard(api, ctx, userId, "Хотите предложить тему для ВКР всем университетам или одному из списка:", keyboard)
	case "add_intern_name":
		userTempData[userId]["intern_name"] = text
		userStates[userId] = "add_intern_req"
		sendMessage(api, ctx, userId, "Введите пререквизиты:")
	case "add_intern_req":
		userTempData[userId]["intern_req"] = text
		userStates[userId] = "add_intern_places"
		sendMessage(api, ctx, userId, "Укажите количество мест:")
	case "add_intern_places":
		places, err := strconv.Atoi(text)
		if err != nil || places <= 0 {
			sendMessage(api, ctx, userId, "Invalid number. Try again.")
			return
		}
		userTempData[userId]["intern_places"] = text
		userStates[userId] = "add_intern_target"
		keyboard := buildTargetKeyboard(api, "intern_target_")
		sendMessageWithKeyboard(api, ctx, userId, "Хотите предложить практику всем университетам или одному из списка:", keyboard)
	default:
		sendMessage(api, ctx, userId, "Ошибка, попробуйте еще раз.")
		cleanupUserData(userId)
		kb := buildStartKeyboard(api, userId)
		sendMessageWithKeyboard(api, ctx, userId, "Menu:", kb)
	}
}

func handleCallback(api *maxbot.Api, ctx context.Context, upd *schemes.MessageCallbackUpdate) {
	userId := upd.Message.Recipient.UserId
	payload := upd.Callback.Payload

	switch {
	case payload == "role_company":
		userStates[userId] = "reg_company_name"
		ensureTempData(userId)
		sendMessage(api, ctx, userId, "Введите название Компании:")
	case payload == "role_uni":
		userStates[userId] = "reg_uni_name"
		ensureTempData(userId)
		sendMessage(api, ctx, userId, "Введите название Университета:")
	case payload == "role_student":
		keyboard := buildUniListKeyboard(api, "student_uni_")
		if keyboard == nil {
			sendMessage(api, ctx, userId, "В системе еще нет ВУЗов.")
			return
		}
		sendMessageWithKeyboard(api, ctx, userId, "Выберите свой ВУЗ:", keyboard)
	case strings.HasPrefix(payload, "student_uni_"):
		uniName := strings.TrimPrefix(payload, "student_uni_")
		if _, ok := unis[uniName]; !ok {
			sendMessage(api, ctx, userId, "Некорректный ВУЗ.")
			return
		}
		ensureTempData(userId)
		userTempData[userId]["uni"] = uniName
		userStates[userId] = "student_course"
		sendMessage(api, ctx, userId, "Введите номер курса:")
	case payload == "add_topic":
		if !isCompany(userId) {
			sendMessage(api, ctx, userId, "Доступно только для компаний.")
			return
		}
		userStates[userId] = "add_topic_name"
		ensureTempData(userId)
		sendMessage(api, ctx, userId, "Введите название темы для ВКР:")
	case payload == "add_intern":
		if !isCompany(userId) {
			sendMessage(api, ctx, userId, "Доступно только для компаний.")
			return
		}
		userStates[userId] = "add_intern_name"
		ensureTempData(userId)
		sendMessage(api, ctx, userId, "Введите заголовок для практики:")
	case payload == "accept_topics":
		uniName := getUniName(userId)
		if uniName == "" {
			sendMessage(api, ctx, userId, "Доступно только для ВУЗов.")
			return
		}
		keyboard := buildAcceptTopicsKeyboard(api, uniName)
		if keyboard == nil {
			sendMessage(api, ctx, userId, "Нет тем ВКР для согласования.")
			return
		}
		sendMessageWithKeyboard(api, ctx, userId, "Выберите темы ВКР для согласования из списка:", keyboard)
	case payload == "accept_interns":
		uniName := getUniName(userId)
		if uniName == "" {
			sendMessage(api, ctx, userId, "Доступно только для ВУЗов.")
			return
		}
		keyboard := buildAcceptInternsKeyboard(api, uniName)
		if keyboard == nil {
			sendMessage(api, ctx, userId, "Нет тем практик для согласования.")
			return
		}
		sendMessageWithKeyboard(api, ctx, userId, "Выберите практики для согласования из списка:", keyboard)
	case payload == "apply_diploma":
		if !isStudent(userId) {
			sendMessage(api, ctx, userId, "Доступно только для студентов.")
			return
		}
		uniName := students[userId].UniName
		keyboard := buildApplyDiplomaKeyboard(api, uniName)
		if keyboard == nil {
			sendMessage(api, ctx, userId, "Нет доступных тем ВКР для вашего ВУЗа.")
			return
		}
		sendMessageWithKeyboard(api, ctx, userId, "Выберите тему ВКР из списка ниже:", keyboard)
	case payload == "apply_intern":
		if !isStudent(userId) {
			sendMessage(api, ctx, userId, "Доступно только для студентов.")
			return
		}
		uniName := students[userId].UniName
		keyboard := buildApplyInternKeyboard(api, uniName)
		if keyboard == nil {
			sendMessage(api, ctx, userId, "Нет доступных практик для вашего ВУЗа.")
			return
		}
		sendMessageWithKeyboard(api, ctx, userId, "Выберите практику из списка ниже:", keyboard)
	case payload == "view_pending_company":
		if !isCompany(userId) {
			return
		}
		inn := getCompanyINN(userId)
		var pending []Application
		for _, app := range applications {
			if app.CompanyINN == inn && app.Status == "pending" {
				pending = append(pending, app)
			}
		}
		if len(pending) == 0 {
			sendMessage(api, ctx, userId, "Нет заявок.")
			return
		}
		sendMessage(api, ctx, userId, "Заявки в процессе:")
		kb := api.Messages.NewKeyboardBuilder()
		for _, app := range pending {
			student := students[app.StudentID]
			itemName := ""
			if app.Type == "diploma" {
				if topic := findDiplomaTopicByID(app.ItemID); topic != nil {
					itemName = topic.Name
				}
			} else {
				if intern := findInternByID(app.ItemID); intern != nil {
					itemName = intern.Name
				}
			}
			label := student.FIO + " for " + app.Type + " '" + itemName + "'"
			kb.AddRow().AddCallback("Accept: "+label, schemes.POSITIVE, "company_accept_"+app.Type+"_"+strconv.Itoa(app.ID))
			kb.AddRow().AddCallback("Reject: "+label, schemes.NEGATIVE, "company_reject_"+strconv.Itoa(app.ID))
		}
		sendMessageWithKeyboard(api, ctx, userId, "", kb)
	case payload == "view_pending_uni":
		if !isUni(userId) {
			return
		}
		uniName := getUniName(userId)
		var pending []Application
		for _, app := range applications {
			if app.UniName == uniName && app.Status == "company_accepted" {
				pending = append(pending, app)
			}
		}
		if len(pending) == 0 {
			sendMessage(api, ctx, userId, "Нет заявок на согласование.")
			return
		}
		sendMessage(api, ctx, userId, "Pending approvals:")
		kb := api.Messages.NewKeyboardBuilder()
		for _, app := range pending {
			student := students[app.StudentID]
			itemName := ""
			if app.Type == "diploma" {
				if topic := findDiplomaTopicByID(app.ItemID); topic != nil {
					itemName = topic.Name
				}
			} else {
				if intern := findInternByID(app.ItemID); intern != nil {
					itemName = intern.Name
				}
			}
			label := student.FIO + " for " + app.Type + " '" + itemName + "'"
			kb.AddRow().AddCallback("Approve: "+label, schemes.POSITIVE, "uni_approve_"+app.Type+"_"+strconv.Itoa(app.ID))
			kb.AddRow().AddCallback("Reject: "+label, schemes.NEGATIVE, "uni_reject_"+strconv.Itoa(app.ID))
		}
		sendMessageWithKeyboard(api, ctx, userId, "", kb)
	case strings.HasPrefix(payload, "target_"):
		target := strings.TrimPrefix(payload, "target_")
		inn := getCompanyINN(userId)
		if inn == "" {
			return
		}
		topic := DiplomaTopic{
			ID:          nextTopicID,
			Name:        userTempData[userId]["topic_name"],
			Description: userTempData[userId]["topic_desc"],
			CompanyINN:  inn,
			TargetUni:   target,
			Status:      "available",
		}
		nextTopicID++
		if target == "general" {
			diplomaTopics = append(diplomaTopics, topic)
		} else {
			specificDiplomaTopics[target] = append(specificDiplomaTopics[target], topic)
			notifyUni(api, ctx, target, "Новая заявка на ВКР: "+topic.Name)
		}
		cleanupUserData(userId)
		sendMessage(api, ctx, userId, "Тема для ВКР согласована.")
	case strings.HasPrefix(payload, "intern_target_"):
		target := strings.TrimPrefix(payload, "intern_target_")
		inn := getCompanyINN(userId)
		if inn == "" {
			return
		}
		places, _ := strconv.Atoi(userTempData[userId]["intern_places"])
		intern := Internship{
			ID:           nextInternID,
			Name:         userTempData[userId]["intern_name"],
			Requirements: userTempData[userId]["intern_req"],
			Places:       places,
			CompanyINN:   inn,
			TargetUni:    target,
			Status:       "available",
		}
		nextInternID++
		if target == "general" {
			internshipOffers = append(internshipOffers, intern)
		} else {
			specificInternships[target] = append(specificInternships[target], intern)
			notifyUni(api, ctx, target, "Новое предложение о практике: "+intern.Name)
		}
		cleanupUserData(userId)
		sendMessage(api, ctx, userId, "Практика добавлена.")
	case strings.HasPrefix(payload, "accept_topic_"):
		idStr := strings.TrimPrefix(payload, "accept_topic_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		uniName := getUniName(userId)
		if uniName == "" {
			return
		}
		uniDiplomaPools[uniName] = append(uniDiplomaPools[uniName], id)
		sendMessage(api, ctx, userId, "ВКР "+idStr+" добавлена к списку тем.")
	case strings.HasPrefix(payload, "accept_intern_"):
		idStr := strings.TrimPrefix(payload, "accept_intern_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		uniName := getUniName(userId)
		if uniName == "" {
			return
		}
		uniInternshipPools[uniName] = append(uniInternshipPools[uniName], id)
		sendMessage(api, ctx, userId, "Практика "+idStr+" добавлена к списку практик.")
	case strings.HasPrefix(payload, "apply_diploma_"):
		idStr := strings.TrimPrefix(payload, "apply_diploma_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		topic := findDiplomaTopicByID(id)
		if topic == nil || topic.Status != "available" {
			sendMessage(api, ctx, userId, "Недоступная тема.")
			return
		}
		app := Application{
			ID:         nextAppID,
			StudentID:  userId,
			Type:       "diploma",
			ItemID:     id,
			Status:     "pending",
			CompanyINN: topic.CompanyINN,
			UniName:    students[userId].UniName,
		}
		nextAppID++
		applications = append(applications, app)
		notifyCompanyForApp(api, ctx, &app)
		sendMessage(api, ctx, userId, "Заявка в Компанию отправлена.")
	case strings.HasPrefix(payload, "apply_intern_"):
		idStr := strings.TrimPrefix(payload, "apply_intern_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		intern := findInternByID(id)
		if intern == nil || intern.Status != "available" || intern.Places <= 0 {
			sendMessage(api, ctx, userId, "Практика недоступна.")
			return
		}
		app := Application{
			ID:         nextAppID,
			StudentID:  userId,
			Type:       "intern",
			ItemID:     id,
			Status:     "pending",
			CompanyINN: intern.CompanyINN,
			UniName:    students[userId].UniName,
		}
		nextAppID++
		applications = append(applications, app)
		notifyCompanyForApp(api, ctx, &app)
		sendMessage(api, ctx, userId, "Заявка в Компанию отправлена.")
	case strings.HasPrefix(payload, "company_accept_diploma_"), strings.HasPrefix(payload, "company_accept_intern_"):
		var prefix string
		if strings.HasPrefix(payload, "company_accept_diploma_") {
			prefix = "company_accept_diploma_"
		} else {
			prefix = "company_accept_intern_"
		}
		idStr := strings.TrimPrefix(payload, prefix)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		app := findAppByID(id)
		if app == nil || app.Status != "pending" {
			return
		}
		app.Status = "company_accepted"
		notifyUniForApp(api, ctx, app)
		sendMessage(api, ctx, userId, "Компания приняла заявку. Уведомление направлено в ваш ВУЗ.")
	case strings.HasPrefix(payload, "company_reject_"):
		idStr := strings.TrimPrefix(payload, "company_reject_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		app := findAppByID(id)
		if app == nil || app.Status != "pending" {
			return
		}
		app.Status = "rejected"
		sendMessage(api, ctx, app.StudentID, "Компания отклонила вашу заявку.")
		sendMessage(api, ctx, userId, "Application rejected.")
	case strings.HasPrefix(payload, "uni_approve_diploma_"):
		idStr := strings.TrimPrefix(payload, "uni_approve_diploma_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		app := findAppByID(id)
		if app == nil || app.Status != "company_accepted" {
			return
		}
		app.Status = "uni_approved"
		if topic := findDiplomaTopicByID(app.ItemID); topic != nil {
			topic.Status = "in work"
		}
		sendMessage(api, ctx, app.StudentID, "Тема ВКР согласована ВУЗом.")
		sendMessage(api, ctx, userId, "Заявка согласована.")
	case strings.HasPrefix(payload, "uni_approve_intern_"):
		idStr := strings.TrimPrefix(payload, "uni_approve_intern_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		app := findAppByID(id)
		if app == nil || app.Status != "company_accepted" {
			return
		}
		app.Status = "uni_approved"
		if intern := findInternByID(app.ItemID); intern != nil {
			intern.Places--
			if intern.Places <= 0 {
				intern.Status = "in work"
			}
		}
		sendMessage(api, ctx, app.StudentID, "Ваша пратика согласована ВУЗом.")
		sendMessage(api, ctx, userId, "Заявка согласована.")
	case strings.HasPrefix(payload, "uni_reject_"):
		idStr := strings.TrimPrefix(payload, "uni_reject_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		app := findAppByID(id)
		if app == nil || app.Status != "company_accepted" {
			return
		}
		app.Status = "rejected"
		sendMessage(api, ctx, app.StudentID, "Заявка отклонена вашим ВУЗом.")
		sendMessage(api, ctx, userId, "Заявка отклонена.")
	}
}
