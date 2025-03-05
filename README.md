# apibase
A batteries-included opinionated API framework and collection of generally useful packages which can also be used standalone


## Opinions
- echo HTTP Server
- OpenAPI REST & gRPC w/ SwaggerUI
- or alternatively HTMX with go templates (server side generated html like php)
- PostgreSQL Database driver wrapper with quality of live improvements (pgx & gopsql)
- Custom user authentication & authorization and basic user management workflow
- API Integration driver for other backend services
- (planned) support for more databases and data source
- (planned) LDAP/SAML user integration
- (planned) built-in admin dashboard


## Packages
Some packages can be used standalone in different projects, these are listed below
- `cli`: for command line interactions, like input dialog
- `log`: logging and custom error handling
- `sqlite`: fully featured sqlite driver w/ scanner to and from typed struct

### cli
Various helper functions for user interaction in the terminal

### logger
(planned) Custom logger function


## Usage
The following is a basic example of how apibase can be used to create an api framework.
```
// TBD
```

### Application Setup
ApiBase serves static files or forwards via reverse proxy any requests that are made, except for those that have a url path starting with `/auth` or `/api`. Other than that any path may be used by the application.

### Authentication
ApiBase provides full user authentication using local auth and/or OAuth (github.com/markbates/goth). In both cases JWT Refresh and Access Tokens are set as http only cookies. Custom access token claim data may be registered by using the `(*web.ApiServer).RegisterAccessClaimDataFunc()` function. In your own api routes, these can be retrieved using the `web.GetAccessClaims()` generic function where data argument is required to be an initialized empty struct of the desired custom claim data.

### Database
In order to add your own apibase database tables, the user must create a sql query and the corresponding struct themselves. Currently, no error-free postgres struct gen library exists that provides the desired functionality. Since this is a one off process in many cases and has horrible rammifications if done incorrectly, a rather manual process is chosen to create a struct for a table and to migrate an existing database table to conform to the updated sql/struct. However, the create sql statement and struct are compared to the current database table which verifies that they match. This is a good middleground and guarantees a stable database interface.

You might be tempted to use an ORM or "advanced" scanning and valuer library, however this is greatly discouraged. It might seem to reduce complexity and therefore developer efficiency, however the added abstractions might bring it's own pitfalls. Writing raw sql and then scanning to a struct (apibase uses [this](https://pkg.go.dev/github.com/georgysavva/scany/v2/pgxscan) library) is quite elegant in it's own right. The same is true for using an orm or valuer library to directly use a struct in a create or update sql query. These might produce nasty side effects, such as updating a default value row with a "uninitialized" (default value zero) element of a struct (e.g. id = 0, created_at = unix time 0)

#### Own Tables
It is not possible to change the built-in tables (users, user_roles, refresh_tokens), however, it is very easy to add additional information to a user by using the users.id foreign key. There are some pgx scan libraries that claim to support scanning nested structs from join queries, however none of them seem to be stable. Even so, a foreign key should be used, since this is a database best practice. To achieve something similar to a join, use database transactions.

## Todo
- [x] Write helper packages to get other projects of the ground first
- [ ] Actually implement the stated goals above
- [ ] Implement the planned features


## Contributions
are very welcome. However, before creating a pull request, please open a detailed issue first, so the exact implementation can be discussed.