```typescript
export interface FullComplexObject {
  /** Object title | Validation: required,min=5,max=100 */
  title: string;
  /** Count | Validation: min=0,max=1000 */
  count: number;
  /** Price | Validation: min=0 */
  price: number;
  /** Is Active */
  active: boolean;
  /** Metadata map */
  metadata: { [key: string]: any };
  /** Nested parent object */
  nested: NestedParent;
  /** Mixed array */
  mixed_array: any[];
  /** Array of objects */
  object_array: { [key: string]: any }[];
  /** 2D array of NestedChild */
  extra_nested: NestedChild[][];
}
```

```typescript
export interface NestedParent {
  /** Parent name | Validation: required,min=3,max=50 */
  name: string;
  /** Creation date */
  created_at: string;
  /** Optional description */
  description: string | null;
  /** Children list */
  children: NestedChild[];
  /** Extra stuff */
  extras: any[];
}
```

```typescript
export interface NestedChild {
  /** Child ID | Validation: min=1,max=9999 */
  id: number;
  /** Tag list */
  tags: string[];
  /** Flags */
  flags: boolean[];
}
```

## User and Friend (Circular Example)

```typescript
export interface User {
  /** User name */
  name: string;
  /** List of friends */
  friends: Friend[];
}
```

```typescript
export interface Friend {
  /** Friend name | Validation: oneof=best close acquaintance | Enum: best,close,acquaintance */
  name: "best" | "close" | "acquaintance";
  /** User reference */
  user: User | null;
}
```

## FruitBasket (Enum/Oneof Example)

```typescript
export interface FruitBasket {
  /** Type of fruit | Validation: oneof=apple banana orange | Enum: apple,banana,orange */
  fruit: "apple" | "banana" | "orange";
  /** Basket size | Validation: oneof=small medium large */
  size: "small" | "medium" | "large";
}
```
