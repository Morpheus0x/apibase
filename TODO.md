
# TODO
- [x] get jwt token from claims by method on the claims struct pointer
- [x] impl local signup
- [x] rename all db structs to singular (remove trailing "s")
- [x] app root static file path or forward port
- [x] Rework log.Error to be a struct compatible with the built-in error interface
- [x] test full auth flow for apibase
- [x] all api responses should return an error code that is translated to a error string client side
- [x] Add feature to save and retrieve scheduled tasks to database
- [x] make separate CSRF middleware completely separate from jwt to also protect login form/the whole api
- [x] Store user agent in refresh_tokens db table entry
- [x] Create new Org for every new user created with admin perms, unless SignupDefaultRoleHook defines own role
- [x] edit auth hooks to modify role? - Done with using an additional hook which may create roles
- [x] Assign a scheduled task to an org
- [x] Protect any referrer uri content by limiting its size to protect against dos and make sure the uri is always starting with the app uri
- [x] fix TODO "refresh JWT": web_oauth/echo_oauth.go#L48
- [x] Resolve all `TODO: remove hardcoded timeout`
- [ ] Refactor all database IDs to be UUIDs
- [ ] Refactor organizations table to to tenants and add "tenant_type" and additional filed for custom "id" for that type
- [ ] Refactor all database tables to impl soft delete, add a deleted column or duplicate all tabels with _deleted suffix (issue with delted column: change all db queries to filter on only not deleted records, what do do with UNIQUE constraints?)
- [ ] check that at least one login type local_auth or oauth_enabled is set to true
- [ ] honor allow_registration flag
- [ ] add language column to users table (maybe not do this!)
- [ ] impl email signup integration, if require_confirmed_email is set, make sure a hook exists that sends the email with user defined template
- [ ] impl email reset api endpoint, add table transient_pw used to store password reset id and expiry time
- [ ] add auth hooks to oauth
- [ ] Make sure that every function that returns an error doens't add any details not also available outside of itself (e.g. userId func param)
- [ ] add captcha for local login and signup
- [ ] check what happens if local auth already exists for user trying to login with same email via oauth
- [ ] impl saml auth
- [ ] impl 2FA, add table 2fa_transient used for 2fa creation and change, lookup how to encrypt 2fa secrets (must be seperate from token_secret to not invalidate 2fa on secret change, DEK & KEK - see [owasp](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html) for details)
- [ ] add parsing on signup username to contain only valid chars and password to not contain spaces (better to have a list of allowed chars)
- [ ] impl htmx api responses

# Old TODOs

## TODO Database Integration
use go generate to import the .sql text file database tables, then use a state machine parser to consume the sql and compare it to the actually existing database table. Also during this comparison, the table struct is checked, if all columns are also present there (tag must match with column name) If the table doesn't exist at all, create it by just running the create query. If the table and corresponding struct is changed in any way, manual table migration needs to be done. 
https://www.startpage.com/do/dsearch?q=golang+include+textfile+in+build+process&cat=web&language=english
https://stackoverflow.com/questions/17796043/how-to-embed-files-into-go-binaries

old:
Add code that verifies that the database contains the expected tables and they match with the struct that will be used to scan the result.
https://www.phind.com/search?cache=c7pgxsam2aelbfl43gta28zd

## TODO OAuth
Look at  
- https://github.com/markbates/goth/blob/master/examples/main.go
- https://www.reddit.com/r/golang/comments/1cf8mji/is_there_a_clear_example_for_using_goth_with_echo/
- https://go.dev/play/p/-RtLSPL4Wsj
to understand how goth does oauth authentication.  
Add code to add support for goth. The current login user flow can be reused, since a user entry in the db needs to be created regardless of signup method.  
After goth returns the user object:
- lookup that user in the db
- skip the password auth (but verify oauth provider callback req).  
- still generate the jwt with my custom claims
Make sure that if the user uses a different oauth provider which returns an email address that already exists in the db, associate that provider with the user (does this even need to be stored in the db, other than "local" or "oauth"?). What if the user later uses the same email returned by oauth to login normally?

## TODO Important
- [x] make sure that the client is always returned to the page from where they clicked login, use the oauth state query param (https://auth0.com/docs/secure/attack-protection/state-parameters)
- [x] have package specific errors always defined in errors.go file inside said package, with an Init() func register those errors with the log package. This is needed if apibase is used with an external program that has their own errors that need to be compareable, maybe by passing the error type to the ErrorNew func, e.g. func ErrorNew\[T myerrtype\](err T, format string, a ...any)
- [x] User [BuntDB](https://github.com/tidwall/buntdb) to store invalidated jwt login tokens
- [ ] Add Support for Hashicorp Vault secret management
- [ ] Add Support for Secret Key Rotation https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html#272-rotation
- [ ] Add Support for 2FA w/ encrypted via DEK and KEK https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html#encrypting-stored-keys

## TODO
- [ ] Every database function is a wrapper that catches errors, if that error is due to db timeout, run reconnect function and the re-run db query
- [x] Rebase go.mod and force push to git by using fixed module name
- [x] Use GoogleCloudPlatform/govanityurls instead of direct github
- [ ] Add API version wrapper function for every endpoint, incrementing the version can have multiple effects: add new endpoint, change implementation, mark as deprecated, remove endpoint. When breaking changes are decided, the apibase api will get a new major or minor version (v1 -> v1.1 or v2 ...). The default behavior is to just use implementation of the previous version. Think about a way to mark an endpoint as deprecated, removed or changed.
- [ ] The apibase api should be either gRPC with created SwaggerUI or OpenAPI definition with example implementation from SwaggerUI, smth like that

## Ideas
- [ ] Registration workflow will create a registration cookie which is valid for 14 days or until the email address is confirmed. After registration and if the email wasn't confirmed the user will always be presented with the registration welcome page. That page gives the option to resend the confirmation email and to change the user email, if the user made a mistake
- [ ] Changing the user email will show buttons to resend the confirmation email and a button to cancel the email change which will invalidate the token sent via email
- [ ] Both operations above will create a one time token which have a defined validity duration and will only be processed if the user clicks on the page and the local javascript sends the tokens from the url query params to the actual api endpoint, showing a loading circle which changes into a green checkmark or red x inside depending on success 
- [ ] Add option for external oauth login: github, ...
- [ ] JWT has short validity of 15min up to 12h, a new token will be transparently created on the backend (this is doen from user secret cookie, I think) for every request made (custom client side request handler required), depending on the Device Type and user preference:
    - until window/browser is closed if not selected remember login (browser, how to determine last window closed of specific site or window closed may be detected via local storage cache?)
    - for 6 months if remember login was selected
    - basically indefinitly if end device has secure, encrypted secret store (Android, iOS)
- [ ] App Health Check (/api/v1/health) this tests any database, integration and login provider