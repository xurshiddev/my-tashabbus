# Roles And Permissions

## Planned Roles

- `SUPER_ADMIN`
- `MFY_CHAIRMAN`
- `STREET_LEADER`
- `RESPONSIBLE_PERSON`

These roles are implemented as API constants. Stage 2 adds MFY, street, and street leader assignment scoping. Household and responsible person workflows are deferred to Stage 3.

## Super Admin

Platform-level administrator. In Stage 1, this role can use the minimal user management endpoints. Future stages will allow platform-wide MFY administration.

Stage 2 permissions:
- Manage all MFYs.
- Assign MFY chairmen.
- Manage streets under any MFY.
- Assign or reassign street leaders to any street.

## MFY Chairman

- Create and edit MFY data.
- Add streets.
- Assign street leaders.
- See all street statistics.
- See household progress inside streets.
- Monitor street leader activity.

Stage 2 permissions:
- Access only own MFY where `users.mfy_id` matches the MFY.
- Create, list, update, activate, and deactivate streets inside own MFY.
- Assign or reassign street leaders only for streets inside own MFY.
- Cannot create other MFYs or manage streets outside own MFY.

## Street Leader

- See assigned streets.
- See households inside assigned streets.
- Assign unlimited responsible persons to own streets.
- Assign household ranges to responsible persons.
- See responsible person progress.
- Check vote counts.

Stage 2 permissions:
- View only actively assigned streets.
- Cannot create or update MFYs.
- Cannot create or update streets.
- Cannot assign street leaders.

## Responsible Person

- See only assigned households.
- Edit house number.
- View and edit total number count.
- Enter voted count.
- Update household status.
- Leave notes.
- Mark household for follow-up.

Stage 2 permissions:
- No MFY or street management permissions.
- Can use only current-user auth endpoints until household workflows are introduced.

Permission checks must be enforced by the API.
