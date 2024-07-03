package utils

const (
	AppName = "job-sender"

	SmtpGmailAddress = "smtp.gmail.com"
	SmtpGmailPort    = 587

	TemplatesDir                = "/templates"
	TemplatesBaseName           = "base.html"
	TemplateListName            = "list.html"
	TemplateEditName            = "edit.html"
	TemplateLoginName           = "login.html"
	TemplateRegisterName        = "register.html"
	TemplateConfirmRegisterName = "confirm_registration.html"
	TemplateSomethingWentWrong  = "something_went_wrong.html"

	UserSessionName           = "user-session"
	UserSessionEmailField     = "email"
	UserSessionTokenField     = "token"
	UserSessionIsVerfiedField = "isVerified"
)
