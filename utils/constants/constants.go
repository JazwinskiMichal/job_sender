package utils

const (
	AppName = "job-sender"

	SmtpGmailAddress = "smtp.gmail.com"
	SmtpGmailPort    = 587

	TemplatesDir                = "/templates"
	TemplatesBaseName           = "base.html"
	TemplateLoginName           = "login.html"
	TemplateRegisterName        = "register.html"
	TemplateConfirmRegisterName = "confirm_registration.html"
	TemplateSomethingWentWrong  = "something_went_wrong.html"

	TemplateContractorsListName = "list_contractors.html"
	TemplateContractorsAddName  = "add_contractor.html"
	TemplateContractorsEditName = "edit_contractor.html"

	TemplateGroupEditName = "edit_group.html"

	TemplateOwnerEditName = "edit_owner.html"

	UserSessionName           = "user-session"
	UserSessionEmailField     = "email"
	UserSessionTokenField     = "token"
	UserSessionIsVerfiedField = "isVerified"
	UserSesstionOwnerIdField  = "ownerID"
)
