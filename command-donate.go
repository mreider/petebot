package main

const DONATE_MESSAGE = `*Donating*

You can donate to any of these causes:

*Petebot*
https://donorbox.org/petebot-donations
I built Petebot in honor of my father, Pete, who was a lifelong runner and died of non hodgkins lymphoma in 2015. Donating to Petebot helps me pay for monthly hosting fees,enhancements, maintenance, babysitting, and occasional date nights with my wife, Alison.

*Compassion and Choices*
https://www.compassionandchoices.org/donate/
This group was very helpful to Pete towards the end of his life. They support people through difficult decisions, and educate the public and professions about end-of-life options.

*Leukemia & Lymphoma Society*
https://donate.lls.org/lls/
LLS is dedicated to funding blood cancer research, education and patient services. Their mission is to cure leukemia, lymphoma, Hodgkin's disease and myeloma, and to improve the quality of life of patients and their families.

`

type DonateCommand struct {
}

func (cmd *DonateCommand) Name() string {
	return "donate"
}

func (cmd *DonateCommand) Execute(params []string, message *IncomingSlackMessage, executor *CommandExecutor) (string, error) {
	return DONATE_MESSAGE, nil
}
