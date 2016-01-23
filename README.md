# Pulse
Pulse will read any log file (text document containing logs) and learn it's patterns! If it is reading a static file than it will wait to the end to email you the logs that did not seem to fit in with the rest. But if you are using the API it will email you live what doesn't fit.
## Team
- Josh Suggs [Github](https://github.com/joshua)
- Michael Dropps [Github](https://github.com/michaeldropps)
- Miguel Espinoza [Github](https://github.com/miguelespinoza)
- Will Dixon [Github](https://github.com/dixonwille)

## TODO
- [ ] Create Algorithm
- [x] Read Config values
  - [x] Files
  - [x] Read a log file on Hard Drive
  - [x] Write to a file to save outstanding logs
- [x] Be able to send emails
  - [x] Use Mailgun's API to send emails
  - [x] Validate User's Email using Mailgun
  - [x] Store keys securely
- [ ] Create the API
  - [ ] POST: log string
  - [ ] POST: file
- [ ] Create Webpage
