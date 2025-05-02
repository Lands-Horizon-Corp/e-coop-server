✅ What AutoMigrate does safely:
1. Adds missing columns
2. Adds missing indexes
3. Adds missing constraints


❌ What AutoMigrate does NOT do:
1. Delete unused columns
2. Change column types
3. Rename columns
4. Handle complex schema changes




 Table general-ledger {
    id unique [pk]
    org id [Ref: < org.id]
    branch id [Ref: < org.id]
    Indexes {
        (org id, branch id) [name:"org-branch"]
        id [unique]
    }

 }