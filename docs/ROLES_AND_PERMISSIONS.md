# Roles And Permissions

## Planned Roles

- `SUPER_ADMIN`
- `MFY_CHAIRMAN`
- `STREET_LEADER`
- `RESPONSIBLE_PERSON`

These roles are implemented as API constants in Stage 1. Full MFY, street, household, and assignment permission scoping is deferred to later stages.

## Super Admin

Platform-level administrator. In Stage 1, this role can use the minimal user management endpoints. Future stages will allow platform-wide MFY administration.

## MFY Chairman

- Create and edit MFY data.
- Add streets.
- Assign street leaders.
- See all street statistics.
- See household progress inside streets.
- Monitor street leader activity.

## Street Leader

- See assigned streets.
- See households inside assigned streets.
- Assign unlimited responsible persons to own streets.
- Assign household ranges to responsible persons.
- See responsible person progress.
- Check vote counts.

## Responsible Person

- See only assigned households.
- Edit house number.
- View and edit total number count.
- Enter voted count.
- Update household status.
- Leave notes.
- Mark household for follow-up.

Permission checks will be implemented in future stages and must be enforced by the API.
