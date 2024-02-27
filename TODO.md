
# TODO
- [ ] Rebase go.mod and force push to git by using fixed module name
- [ ] Use GoogleCloudPlatform/govanityurls instead of direct github

# Ideas
- [ ] Registration workflow will create a registration cookie which is valid for 14 days or until the email address is confirmed. After registration and if the email wasn't confirmed the user will always be presented with the registration welcome page. That page gives the option to resend the confirmation email and to change the user email, if the user made a mistake
- [ ] Changing the user email will show buttons to resend the confirmation email and a button to cancel the email change which will invalidate the token sent via email
- [ ] Both operations above will create a one time token which have a defined validity duration and will only be processed if the user clicks on the page and the local javascript sends the tokens from the url query params to the actual api endpoint, showing a loading circle which changes into a green checkmark or red x inside depending on success 
- [ ] Add option for external oauth login: github, ...
- [ ] JWT has short validity of 15min up to 12h, a new token will be transparently created on the backend (this is doen from user secret cookie, I think) for every request made (custom client side request handler required), depending on the Device Type and user preference:
    - until window/browser is closed if not selected remember login (browser, how to determine last window closed of specific site or window closed may be detected via local storage cache?)
    - for 6 months if remember login was selected
    - basically indefinitly if end device has secure, encrypted secret store (Android, iOS)
- [ ] Any base session (not specific JWT token) will be saved and the user has the ability to invalidate a specific session, useragent and other identifying information will be used alogside creation and last usage timestamp
- [ ] App Health Check (/api/v1/health) this tests any database, integration and login provider