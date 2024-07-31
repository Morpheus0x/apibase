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
TBD


## Todo
- [x] Write helper packages to get other projects of the ground first
- [ ] Actually implement the stated goals above
- [ ] Implement the planned features


## Contributions
are very welcome. However, before creating a pull request, please open a detailed issue first, so the exact implementation can be discussed.