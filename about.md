# Single Table Approach
- save all namespace in one table
- no foreign constraint check
-

# Multiple table Approach



## Action
- Viewer
- Creator
- Write
- Apply
- Execute

## SUbject ID
User in question


- for global access (i want to restrict who can create/view a batch change)
ID: 1
Namespace: BatchChange
NamespaceObjectID: *
Relation: View
SubjectID: NULL

Bayo -> Default (role) -> Perm{1}

Ray -> Default (role) -> Perm{1}

5 mins later, i remove 1 from default role. But i want to give Ray alone
access to view batch changes.

Bayo -> Default (role)
Ray -> Default (role)

ID: 1
Namespace: BatchChange
NamespaceObjectID: *
Relation: View
SubjectID: NULL

Create another role called

```sql
SELECT * from permissions p
where
    p.namespace = 'batch_change' | 'notebooks'
        AND
    p.namespace_object_id IS NULL
        AND
    p.relation = 'WRITE'
inner join on user where id = 4
inner join on user_roles where user_id = user.id | 4
inner join on role_permissions where rp.role_id = ur.id
```

ACLs

