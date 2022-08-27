<!-- < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < ☺
v                               ✰  Thanks for creating a PR! ✰
v    Before smashing the submit button please review the checkboxes.
v    If a checkbox is n/a - please still include it but + a little note why
v    If your PR doesn't close an issue, that's OK!  Just remove the Closes: #XXX line!
☺ > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > >  -->

Closes: #XXX

## Context and purpose of the change

<!-- Add a description of the overall background and high level changes that this PR introduces

_(E.g.: This pull request improves documentation of area A by adding ...._

## Brief Changelog

_(for example:)_

- _The metadata is stored in the blob store on job creation time as a persistent artifact_
- _Deployments RPC transmits only the blob storage reference_
- _Daemons retrieve the RPC data from the blob cache_

## Author's Checklist

_(Please pick one of the following options)_

- [ ] Run and PASSED locally all GAIA integration tests
- [ ] If the change is contentful, I either:
    - [ ] Added a new unit test OR 
    - [ ] Added test cases to existing unit tests
- [ ] OR this change is a trivial rework / code cleanup without any test coverage

_(or)_

This change is already covered by existing tests, such as _(please describe tests)_.

_(or)_

I have...

_(example:)_

- _Added unit test that validates ..._
- _Added integration tests for end-to-end deployment with ..._
- _Extended integration test for ..._
- _Manually verified the change by ..._

## Documentation and Release Note

- Does this pull request introduce a new feature or user-facing behavior changes? (yes / no)
- Is a relevant changelog entry added to the `Unreleased` section in `CHANGELOG.md`? (yes / no)
- How is the feature or change documented? (not applicable / specification (`x/<module>/spec/`) / README.md / not documented)
- Does this pull request update existing proto field values (and require a backend and frontend migration)? (yes / no)
- Does this pull request change existing proto field names (and require a frontend migration)? (yes / no)
