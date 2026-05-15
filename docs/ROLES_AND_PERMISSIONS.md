# Roles And Permissions

## Planned Roles

- `SUPER_ADMIN`
- `MFY_CHAIRMAN`
- `STREET_LEADER`
- `RESPONSIBLE_PERSON`

These roles are implemented as API constants. Stage 3 adds household and responsible person assignment scoping.

## Super Admin

Platform-level administrator. In Stage 1, this role can use the minimal user management endpoints. Future stages will allow platform-wide MFY administration.

Stage 2 permissions:
- Manage all MFYs.
- Assign MFY chairmen.
- Manage streets under any MFY.
- Assign or reassign street leaders to any street.

Stage 3 permissions:
- Create, list, view, and update households in any street.
- Assign responsible persons to any street.
- Deactivate responsible assignments.
- View household change logs.

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

Stage 3 permissions:
- Manage households only inside own MFY streets.
- Assign responsible persons only inside own MFY streets.
- View responsible assignments and household change logs inside own MFY.

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

Stage 3 permissions:
- Create, list, view, and update households only in actively assigned streets.
- Assign responsible persons only inside actively assigned streets.
- View responsible assignments and household logs only for assigned streets.

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

Stage 3 permissions:
- View only households assigned to them.
- Update assigned household house number, total numbers, contacted numbers, reported voted numbers, status, and notes.
- Cannot create MFYs, streets, households, or assignments.
- Cannot view other responsible persons' households.

Permission checks must be enforced by the API.
