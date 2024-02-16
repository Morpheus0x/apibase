
# TODO
- [ ] Rebase go.mod and force push to git by using fixed module name
- [ ] Use GoogleCloudPlatform/govanityurls instead of direct github

# Ideas
- [ ] Registration workflow will create a registration cookie which is valid for 14 days or until the email address is confirmed. After registration and if the email wasn't confirmed the user will always be presented with the registration welcome page. That page gives the option to resend the confirmation email and to change the user email, if the user made a mistake
- [ ] Changing the user email will show buttons to resend the confirmation email and a button to cancel the email change which will invalidate the token sent via email
- [ ] Both operations above will create a one time token which have a defined validity duration and will only be processed if the user clicks on the page and the local javascript sends the tokens from the url query params to the actual api endpoint, showing a loading circle which changes into a green checkmark or red x inside depending on success 
- [ ] Add option for external oauth login: github, ...
