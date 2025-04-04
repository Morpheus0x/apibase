# apibase
A batteries-included opinionated API framework and collection of generally useful packages which can also be used standalone


## Opinions
- [toml](https://github.com/BurntSushi/toml) - Config File Format superior to any other
- [cobra](https://github.com/spf13/cobra) - CLI arg parsing
- [pgx](https://github.com/jackc/pgx) and [pgxscan](https://pkg.go.dev/github.com/georgysavva/scany/v2/pgxscan) from [scany](https://github.com/georgysavva/scany/) - PostgreSQL Database driver and scanning to struct
- [echo](https://github.com/labstack/echo) - HTTP Server
- [jwt](https://github.com/golang-jwt/jwt/) - Json Web Token
- [goth](https://github.com/markbates/goth) - OAuth User Authentication (optional)
- REST API (SwaggerUI integration planned)
- Custom user authentication & authorization and basic user management workflow
- (planned) LDAP/SAML user integration
- (planned) support for more databases and data source
- (planned) built-in admin dashboard
- (planned) HTMX and templ (server side generated html like php) in addition to REST API


## apibase Packages
### Commmon
These packages can be used standalone in different projects, regardless if they are built with apibase:
- `cli`: for command line interactions, like input dialog
- `cron`: custom task scheduling (has apibase database integration but can be used standalone)
- `email`: email sending
- `errx`: custom error type with special error chain feature (nested errors)
- `log`: custom logging library
- `sqlite`: fully featured sqlite driver w/ scanner to and from typed struct (refactoring needed)
### apibase specific
All other packages more or less require apibase to be used:
- `base`: entrypoint to configure and setup apibase
- `baseconfig`: apibase global config defines
- `cmd`: cli for apibase
- `db`: database connection and apibase internal db transactions
- `grpc`: not used/implemented
- `helper`: common helper functions used by apibase
- `hook`: can be used to extend the default user auth flow, e.g. custom permissions for new users
- `integrations`: not used/implemented
- `rest`: not used/implemented
- `table`: default apibase db tables and their corresponding structs
- `web`: webserver setup w/ jwt auth
- `web_auth`: local user signup, login and logout functions
- `web_oauth`: OAuth user login/signup and logout functions
- `web_response`: REST API Json Response and common response id defenitions
- `web_setup`: Webserver bringup and built-in api endpoints setup


## Usage
The following is a basic example of how apibase can be used to create an api framework.
```
// TBD
```

### Application Setup
ApiBase serves static files or forwards via reverse proxy any requests that are made, except for those that have a url path starting with `/auth` or `/api`. Other than that any path may be used by the static files or proxied application.

### Authentication
ApiBase provides full user authentication using local auth and/or OAuth (github.com/markbates/goth). In both cases JWT Refresh and Access Tokens are set as http only cookies. Custom access token claim data may be registered by using the `(*web.ApiServer).RegisterAccessClaimDataFunc()` function. In your own api routes, these can be retrieved using the `web.GetAccessClaims()` generic function where data argument is required to be an initialized empty struct of the desired custom claim data.

### Database
In order to add your own apibase database tables, the user must create a sql query and the corresponding struct themselves. Currently, no error-free postgres struct gen library exists that provides the desired functionality. Since this is a one off process in many cases and has horrible rammifications if done incorrectly, a rather manual process is chosen to create a struct for a table and to migrate an existing database table to conform to the updated sql/struct. _However, the create sql statement and struct are compared to the current database table which verifies that they match. This is a good middleground and guarantees a stable database interface. - not yet implemented_

You might be tempted to use an ORM or "advanced" scanning and valuer library, however this is greatly discouraged. It might seem to reduce complexity and therefore developer efficiency, however the added abstractions might bring it's own pitfalls. Writing raw sql and then scanning to a struct (apibase uses pgxscan from the scany library) is quite elegant in it's own right. The same is true for using an orm or valuer library to directly use a struct in a create or update sql query. These might produce nasty side effects, such as updating a default value row with a uninitialized (default "zero" value) element of a struct (e.g. id = 0, created_at = unix time 0)

#### Own Tables
It is not possible to change the built-in tables (users, user_roles, refresh_tokens), however, it is very easy to add additional information to a user by using the users.id foreign key. There are some pgx scan libraries that claim to support scanning nested structs from join queries, however none of them seem to be stable. Even so, a foreign key should be used, since this is a database best practice. Database join queries can still be performed but need special consideration when scanning using scany, alternatively database transactions are recommended to achieve basically the same thing.

## Contributions
are very welcome. However, before creating a pull request, please open a detailed issue first, so the exact implementation can be discussed.