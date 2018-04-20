package main

const (
	UNLOCK_EMAIL_SUBJECT = "Petebot unlock code"
	UNLOCK_EMAIL_BODY    = `Dearest Petebot user,

	Thank you so much for donating to {charity}.

	Your unlock code is {code}

	You can unlock your account by returning to slack and typing:

	@petebot unlock 123456 <--- replace that with your code

	Happy petebotting!

	Matt
`
)
