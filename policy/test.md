```typescript
export interface UUIDExample {
  /** Unique identifier */
  id: number[];
  /** Optional parent ID */
  parent_id: string | null;
  /** List of child IDs */
  children: number[][];
  /** Creation time */
  created_at: string;
  /** Metadata map */
  meta: { [key: string]: any };
}
```

```typescript
export interface ComplexUUID {
  /** Main UUID */
  primary_id: number[];
  /** Secondary UUID */
  secondary_id: string | null;
  /** UUID history */
  history: string | null[];
  /** 2D UUIDs */
  extra: number[][][];
  /** Example UUID struct */
  example: UUIDExample;
}
```
