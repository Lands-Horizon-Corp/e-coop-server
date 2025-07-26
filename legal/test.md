## TypeScript Interfaces for Go Playground Validator Structs

```typescript
export interface UserLoginRequest {
  /**
  Validation: required
*/
  key: string;
  /**
  Validation: required, min: 8
*/
  password: string;
}
```

```typescript
export interface UserRegisterRequest {
  /**
  Validation: required, format: email
*/
  email: string;
  /**
  Validation: required, min: 8
*/
  password: string;
  birthdate?: string;
  /**
  Validation: required, min: 3, max: 100
*/
  user_name: string;
  full_name?: string;
  first_name?: string;
  middle_name?: string;
  last_name?: string;
  suffix?: string;
  /**
  Validation: required, min: 7, max: 20
*/
  contact_number: string;
  media_id?: string;
}
```
