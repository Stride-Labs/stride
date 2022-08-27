<!-- < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < ☺
v                               ✰  Thanks for creating a PR! ✰    
v    Before smashing the submit button please review the checkboxes.
v    If a checkbox is n/a - please still include it but + a little note why
v    If your PR doesn't close an issue, that's OK!  Just remove the Closes: #XXX line!
☺ > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > >  -->

Closes: #XXX

## Context and purpose of the change

<!-- Add a description of the overall background and high level changes that this PR introduces

*(E.g.: This pull request improves documentation of area A by adding ....* -->


## Brief Changelog

<!-- *(for example:)*
 
  - *The metadata is stored in the blob store on job creation time as a persistent artifact*
  - *Deployments RPC transmits only the blob storage reference*
  - *Daemons retrieve the RPC data from the blob cache* -->


## Author's Checklist

I have...

- [ ] Run and PASSED locally all GAIA integration tests
- [ ] If the change is contentful, I either:
    - [ ] Added a new unit test OR 
    - [ ] Added test cases to existing unit tests
- [ ] OR this change is a trivial rework / code cleanup without any test coverage

If skipped any of the tests above, explain.
<!-- *(example:)*
  - *Added unit test that validates ...*
  - *Added integration tests for end-to-end deployment with ...*
  - *Extended integration test for ...*
  - *Manually verified the change by ...* -->

## Reviewers Checklist

*All items are required. Please add a note if the item is not applicable and please add
your handle next to the items reviewed if you only reviewed selected items.*

I have...

- [ ] reviewed state machine logic
- [ ] reviewed API design and naming
- [ ] manually tested (if applicable)
- [ ] confirmed the author wrote unit tests for new logic
- [ ] reviewed documentation exists and is accurate


## Documentation and Release Note

  - [ ] Does this pull request introduce a new feature or user-facing behavior changes? 
  - [ ] Is a relevant changelog entry added to the `Unreleased` section in `CHANGELOG.md`?
  - [ ] This pull request updates existing proto field values (and require a backend and frontend migration)? 
  - [ ] Does this pull request change existing proto field names (and require a frontend migration)?
  How is the feature or change documented? 
      - [ ] not applicable
      - [ ] jira ticket `XXX` 
      - [ ] specification (`x/<module>/spec/`) 
      - [ ] README.md 
      - [ ] not documented <!-- because ... EXPLAIN WHY! -->
