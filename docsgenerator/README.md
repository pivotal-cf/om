# `om` Docs Generator

## What this does

docsgenerator will take the usage defined in the command definition,
and combine it with an optional 
`EXPANDED_DESCRIPTION.md` and optional `ADDITIONAL_INFO.md` 
in order to create a unified `README.md` for the command
that can automatedly be kept up-to-date.

To rephrase or offer a better summary of what the command does,
add information to the `EXPANDED_DESCRIPTION.md`.
In the generated doc, this will come before the command usage.

To add additional documentation to the `README.md` of any command, 
("in addition" to the already existing flag 
and basic command description),
add a file called `ADDITIONAL_INFO.md`
to the appropriate folder in the `templates` directory.
This can include helpful examples of 
how to use a certain flag of the command,
or other helpful information to better understand the command.
Information in this file will be displayed
after the extended description and usage for the command.

## How to use it

To generate docs from the usage and templates, run
```bash
go run docsgenerator/update-docs.go
```
and commit the result. 
The Platform Automation team would prefer 
the docs update to be a separate commit.