package utils

const (
	AppName = "job-sender"
	AppUrl  = "https://app.jobsender.pl"

	SmtpGmailAddress = "smtp.gmail.com"
	ImapGmailAddress = "imap.gmail.com"
	SmtpGmailPort    = 587

	TemplatesDir                = "/templates"
	TemplatesBaseName           = "base.html"
	TemplateLoginName           = "login.html"
	TemplateRegisterName        = "register.html"
	TemplateConfirmRegisterName = "confirm_registration.html"
	TemplateSomethingWentWrong  = "something_went_wrong.html"

	TemplateOwnerAddName  = "add_owner.html"
	TemplateOwnerEditName = "edit_owner.html"

	TemplateGroupAddName  = "add_group.html"
	TemplateGroupEditName = "edit_group.html"

	TemplateContractorsGetName  = "get_contractors.html"
	TemplateContractorsAddName  = "add_contractor.html"
	TemplateContractorsEditName = "edit_contractor.html"

	UserSessionName                = "user-session"
	TimesheetAggegationSessionName = "timesheet-aggregation-session"

	SessionEmailField     = "email"
	SessionTokenField     = "token"
	SessionIsVerfiedField = "isVerified"
	SesstionOwnerIdField  = "ownerID"

	SessionAggregatorIDField        = "aggregatorID"
	SessionLastAggregationTimeField = "lastAggregationTime"
)
