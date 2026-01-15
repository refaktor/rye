# Embracing Failure: Error Handling in Rye (and why try/catch is still the GOTO)

In most programming languages, errors are treated as exceptional control-flow: you jump out of “where you are” to some handler “somewhere else”.

Rye goes the other direction: failures are *values* plus a small amount of *explicit state* (the failure flag / return flag). That combination lets you build higher-intent error handling out of composable combinators like `check`, `fix`, and `ensure`, instead of relying on a “jump to handler” mechanism.

Let's dive into Rye's failure handling system and see how it compares to Python's exception handling, highlighting how Rye's approach can make your code more robust and expressive.

## Quick Reference

| Category | Rye Functions |
|----------|---------------|
| **Creation** | `fail`, `^fail`, `failure`, `refail`, `failure\wrap` |
| **Inspection** | `is-error`, `error-kind?`, `is-error-of-kind`, `cause?`, `status?`, `message?`, `details?`, `has-failed`, `is-failure`, `is-success` |
| **Handling** | `disarm`, `check`, `^check`, `ensure`, `^ensure`, `fix`, `^fix`, `fix\either`, `fix\else`, `fix\continue`, `continue`, `^fix\match` |
| **Control** | `try`, `try-all`, `try\in`, `finally`, `retry`, `persist`, `timeout`, `requires-one-of`, `^requires-one-of` |

> Note: Rye does **not** currently have language-level `try/catch` or `throw`/`raise` keywords in the Go implementation. The closest equivalent to “catch” is *combinator-based recovery* (`fix`, `fix\either`, `fix\continue`) and *boundary blocks* like `try`.

## The Three Faces of Failure

Rye offers three main ways to create errors, each with its own behavior:

```rye
; Creates an error and sets the failure flag (evaluation continues)
fail "Something went wrong"

; Creates an error and sets failure + return flags (exits the current function)
^fail "Something went catastrophically wrong"

; Creates an error value without touching flags
err: failure "This is just an error value"
```

The difference is subtle but powerful. `fail` signals a problem but lets the surrounding code decide what to do, while `^fail` is more decisive, immediately exiting the current function. And `failure` just creates an error value without affecting program flow at all.

### Python comparison (control-flow)

In Python, exceptions are primarily raised and caught:

```python
# Raises an exception that will propagate up unless caught
raise ValueError("Something went wrong")

# Creating an exception without raising it (similar to Rye's 'failure')
err = ValueError("This is just an error value")

# Python has no direct equivalent to Rye's ^fail, but you could simulate it with:
def function_with_early_return():
    # ...
    if error_condition:
        raise ValueError("Something went catastrophically wrong")
    # ...
```

Unlike Rye, Python doesn't distinguish between “make an error value” and “start unwinding”. `raise` immediately triggers a non-local jump to a handler.

## Anatomy of an Error

Errors in Rye aren't just strings; they can carry structured data.

```rye
; message
validation-error: failure "Invalid email format"

; status
api-error: failure 404

; kind + status + message + details
db-error: failure {
  'db-error 1001 "Insert failed"
  "table" "users"
  "operation" "insert"
}
```

You can then inspect these errors using accessor functions:

```rye
api-error |status?           ; Returns 404
validation-error |message?   ; Returns "Invalid email format"
db-error |details? .table    ; Returns "users"
```

### Python comparison

Python's exceptions can also carry additional information, but it's typically done through subclassing or instance attributes:

```python
# Standard exception with a message
validation_error = ValueError("Invalid email format")

# Custom exception with status code
class APIError(Exception):
    def __init__(self, status_code, message=None):
        self.status_code = status_code
        self.message = message or f"API error: {status_code}"
        super().__init__(self.message)

api_error = APIError(404, "Not found")

# Custom exception with multiple attributes
class DBError(Exception):
    def __init__(self, code, table, operation):
        self.code = code
        self.table = table
        self.operation = operation
        message = f"Database error {code} on {table} during {operation}"
        super().__init__(message)

db_error = DBError(1001, "users", "insert")

# Accessing the information
print(api_error.status_code)  # 404
print(validation_error.args[0])  # "Invalid email format"
print(db_error.table)  # "users"
```

Python requires more boilerplate to create structured exceptions, and there's no standardized way to access attributes across different exception types.

## The art of handling failures (combinators, not jumps)

The most important thing to understand is the **failure flag**.

- `fail` sets `FailureFlag = true` and returns an error value.
- `^fail` sets `FailureFlag = true` *and* `ReturnFlag = true` and returns an error value.
- Many “inspector” functions set `FailureFlag = false` so you can inspect errors without propagating them.

This explains why Rye error handling feels like a *data pipeline* and not like `try/catch`: the “error path” is visible in the chain.

Where Rye really shines is in how it lets you handle errors. Let's look at some patterns:

### The `fix` pattern (recover)

`fix` executes its block **only if** the value is in failure state (or is an error value). It also clears the failure flag before running the handler, so the handler starts “clean”.

```rye
get-user 123 |fix { default-user }
parse-date user-input |fix { today }
```

#### Python comparison

Python uses try/except blocks for error recovery:

```python
# Try to get a user, or return a default if it fails
try:
    user = get_user(123)
except Exception:
    user = default_user

# Try to parse a date, or return today if it fails
try:
    date = parse_date(user_input)
except Exception:
    date = today
```

The Python approach is more verbose and doesn't chain as elegantly as Rye's pipe-based approach.

### The `check` pattern (add context)

`check` is for **adding context**.

If the input is in failure state, it wraps the original error with a new error created from the second argument. In the Go implementation, this preserves the original error as a *parent* (error chain).

```rye
db/query "SELECT * FROM users" |check "Error querying users database"

file/read "config.json"
  |check "Failed to read config file"
  |json/parse
  |check "Config file contains invalid JSON"
```

#### Python comparison

Python typically uses exception chaining or re-raising with additional context:

```python
# Add context to any error that might occur
try:
    result = db.query("SELECT * FROM users")
except Exception as e:
    raise RuntimeError("Error querying users database") from e

# In a chain of operations, add context at each step
try:
    with open("config.json", "r") as f:
        config_text = f.read()
except Exception as e:
    raise RuntimeError("Failed to read config file") from e

try:
    config = json.loads(config_text)
except Exception as e:
    raise RuntimeError("Config file contains invalid JSON") from e
```

Python's exception chaining with `raise ... from` preserves the original exception, but it requires more boilerplate and doesn't compose as naturally in a data pipeline.

### The `ensure` pattern (validate)

`ensure` turns boolean-ish checks into a failure value when they’re falsey.

```rye
; Validate user input
user-age > 0 |ensure "Age must be positive"
user-email |contains "@" |ensure "Email must contain @"

; Check preconditions before proceeding
db/connected? |ensure "Database connection required"
user/has-permission? 'admin |ensure "Admin permission required"
```

#### Python Comparison

Python typically uses assertions or explicit conditionals with exceptions:

```python
# Validate user input
if not user_age > 0:
    raise ValueError("Age must be positive")
if "@" not in user_email:
    raise ValueError("Email must contain @")

# Check preconditions before proceeding
if not db.connected:
    raise ConnectionError("Database connection required")
if not user.has_permission("admin"):
    raise PermissionError("Admin permission required")

# Or using assert (though not recommended for production code)
assert user_age > 0, "Age must be positive"
assert "@" in user_email, "Email must contain @"
```

Python requires more verbose conditional checks and explicit exception raising, compared to Rye's more declarative approach.

## Examples from the repository

The repo includes a runnable demo: `examples/latest_to_migrate/error_handling_demo.rye`.

Key patterns shown there:

- **Try + fix**

```rye
result2: try { divide-safely 10 0 |fix { "Used default value instead" } }
```

- **Check + continue** (add context, then keep going only on success)

```rye
process-user user-id
  |check "Error processing user request"
  |continue { user ->
      print ["Processing user:" user.name]
      user
  }
```

- **Ensure**

```rye
age > 0 |ensure "Age must be positive"
age >= 18 |ensure "Must be at least 18 years old"
```

- **finally** for cleanup without swallowing the original failure

```rye
finally { fail "Operation failed" } { resource:: "closed" }
```

Let's look at some practical examples of how these patterns come together:

### Example 1: API Request with Error Handling

```rye
fetch-user-data: fn { user-id } {
  ; Validate input
  user-id > 0 |^ensure "User ID must be positive"
  
  ; Try to fetch the user, with multiple layers of error handling
  http/get (join "https://api.example.com/users/" user-id)
    |check "API request failed" 
    |json/parse 
    |check "Invalid JSON response"
    |fix { 
      ; If anything failed, return a default user
      { "id" user-id "name" "Unknown" "status" "error" }
    }
}
```

#### Python Comparison

```python
def fetch_user_data(user_id):
    # Validate input
    if not user_id > 0:
        raise ValueError("User ID must be positive")
    
    # Try to fetch the user, with multiple layers of error handling
    try:
        try:
            response = requests.get(f"https://api.example.com/users/{user_id}")
            response.raise_for_status()  # Raises HTTPError for bad responses
        except requests.RequestException as e:
            raise RuntimeError("API request failed") from e
        
        try:
            return response.json()
        except ValueError as e:
            raise RuntimeError("Invalid JSON response") from e
    except Exception:
        # If anything failed, return a default user
        return {
            "id": user_id,
            "name": "Unknown",
            "status": "error"
        }
```

The Python version requires nested try/except blocks and explicit re-raising of exceptions with context, making it more verbose and harder to follow than the Rye version.

### Example 2: Database Transaction with Retry

```rye
save-user: fn { user } {
  ; Retry the database operation up to 3 times
  retry 3 {
    ; Ensure we have a valid user
    user/valid? user |^ensure "Invalid user data"
    
    ; Try to save, with timeout protection
    timeout 5000 {
      db/begin-transaction
      
      ; Use defer to ensure transaction is rolled back on any failure
      defer { db/rollback-transaction }
      
      db/insert "users" user
        |^check "Failed to insert user"
      
      db/commit-transaction
        |^check "Failed to commit transaction"
      
      "User saved successfully"
    }
  }
}
```

#### Python Comparison

```python
def save_user(user):
    # Retry the database operation up to 3 times
    for attempt in range(3):
        try:
            # Ensure we have a valid user
            if not user_valid(user):
                raise ValueError("Invalid user data")
            
            # Try to save, with timeout protection
            try:
                # Set up a timeout
                signal.alarm(5)  # 5 seconds timeout
                
                try:
                    db.begin_transaction()
                    
                    try:
                        db.insert("users", user)
                    except Exception as e:
                        raise RuntimeError("Failed to insert user") from e
                    
                    try:
                        db.commit_transaction()
                    except Exception as e:
                        raise RuntimeError("Failed to commit transaction") from e
                    
                    return "User saved successfully"
                except Exception:
                    # Ensure transaction is rolled back on any failure
                    db.rollback_transaction()
                    raise
                finally:
                    # Cancel the timeout
                    signal.alarm(0)
            except TimeoutError:
                continue  # Try again if timed out
        except Exception:
            if attempt == 2:  # Last attempt
                raise  # Re-raise the exception if all retries failed
            continue  # Otherwise try again
```

The Python version is significantly more complex, with nested try/except blocks, manual timeout handling, and explicit retry logic. The Rye version is much more concise and readable, with built-in support for retries, timeouts, and deferred cleanup.

### Example 3: Cascading Error Handling

```rye
process-order: fn { order } {
  ; A chain of operations where any step can fail
  validate-order order
    |fix\continue {
      ; Error handler
      log/error "Order validation failed"
      fail "Invalid order"
    } {
      ; Success handler - continues the chain
      check-inventory order
        |fix\continue {
          log/error "Inventory check failed"
          fail "Items out of stock"
        } {
          process-payment order
            |fix\continue {
              log/error "Payment processing failed"
              fail "Payment declined"
            } {
              ship-order order
                |fix {
                  log/error "Shipping failed"
                  fail "Shipping error"
                }
            }
        }
    }
}
```

#### Python Comparison

```python
def process_order(order):
    # A chain of operations where any step can fail
    try:
        validated_order = validate_order(order)
    except Exception:
        logging.error("Order validation failed")
        raise ValueError("Invalid order")
    
    try:
        inventory_result = check_inventory(validated_order)
    except Exception:
        logging.error("Inventory check failed")
        raise ValueError("Items out of stock")
    
    try:
        payment_result = process_payment(validated_order)
    except Exception:
        logging.error("Payment processing failed")
        raise ValueError("Payment declined")
    
    try:
        return ship_order(validated_order)
    except Exception:
        logging.error("Shipping failed")
        raise ValueError("Shipping error")
```

The Python version uses a series of try/except blocks, which is more straightforward than the Rye version in this case. However, it doesn't provide the same level of control over the error flow - in Rye, each step can decide whether to propagate the error, transform it, or recover from it, while in Python, each exception immediately jumps to the corresponding except block.

## Semantics notes (important details)

These are implementation-aligned details (from `evaldo/builtins_error_handling.go`):

1. **`try` is a boundary**
   - It runs a block with `InErrHandler = true`.
   - It clears `ReturnFlag`, `ErrorFlag`, and `FailureFlag` after evaluation.
   - It does *not* print failures while inside the `try` builtin.

   This means `try { ... }` is closer to “evaluate and return whatever value you got, but don’t leak failure state outward”.

2. **`fix` clears the failure flag before executing its handler**
   - So your handler starts in a non-failed state, and can choose to `fail` again.

3. **Inspector functions typically clear the failure flag**
   - Example: `is-error`, `error-kind?`, `has-failed` set `FailureFlag = false` so inspection doesn’t keep you in failure state.

4. **`continue` is the dual of `fix`**
   - It runs only on success.
   - On failure it returns the original value and clears failure state (so you can still inspect it).

## The philosophy of failure

Rye's approach to error handling reflects a deeper philosophy: failures are normal, expected parts of a program's execution. Rather than treating them as exceptional cases to be avoided, we embrace them as values that can be passed around, transformed, and handled just like any other data.

This leads to code that's more robust, more expressive, and often more concise than traditional error handling approaches. Instead of deeply nested try/catch blocks or error codes that must be checked after every operation, Rye lets you express your error handling logic as a natural part of your data flow.

### Key Differences from Python

1. **Errors as Values**: In Rye, errors are just values that can be passed around, while in Python, exceptions are special control flow mechanisms.

2. **Composability**: Rye's error handling combinators compose naturally with its pipe-based syntax, while Python's try/except blocks don't compose well.

3. **Granular Control**: Rye gives you fine-grained control over when to propagate, transform, or recover from errors, while Python's exceptions are more all-or-nothing.

4. **Explicit Error Paths**: In Rye, error handling is explicit in the data flow, making it easier to see how errors are handled, while Python's exception handling can be more implicit and harder to follow.

5. **Built-in Patterns**: Rye provides built-in combinators for common error handling patterns, while Python requires you to implement these patterns yourself.

So next time you're writing Rye code: don’t *catch everything*; instead, structure error handling with intent.

---

## Complete function reference (aligned with builtins)

### Error Creation

| Function | Description | Example |
|----------|-------------|---------|
| `fail` | Creates error, sets failure flag, continues | `fail "error message"` |
| `^fail` | Creates error, sets failure flag, returns immediately | `^fail 404` |
| `failure` | Creates error object without setting any flags | `err: failure "just a value"` |
| `refail` | Re-raises error with additional context | `refail old-err "new context"` |
| `failure\wrap` | Creates error wrapping another error | `failure\wrap "outer" inner-err` |

### Error Inspection

| Function | Description | Example |
|----------|-------------|---------|
| `is-error` | Returns true if value is an error | `is-error val` |
| `error-kind?` | Returns error's kind as word, or void | `error-kind? err` |
| `is-error-of-kind` | Checks if error is of specific kind | `is-error-of-kind err 'not-found` |
| `cause?` | Extracts root cause from error chain | `cause? wrapped-err` |
| `status?` | Extracts numeric status code | `status? err` → `404` |
| `message?` | Extracts message string | `message? err` → `"error"` |
| `details?` | Extracts additional details as dict | `details? err` |
| `has-failed` | Tests if value is error (accepts failures) | `has-failed val` |
| `is-failure` | Checks if value is error type | `is-failure val` |
| `is-success` | Returns true for non-error values | `is-success val` |

### Error Handling

| Function | Description | Example |
|----------|-------------|---------|
| `disarm` | Clears failure flag, preserves error object | `err \|disarm` |
| `check` | Wraps error with context if failed | `val \|check "context"` |
| `^check` | Like check, but also returns immediately | `val \|^check "context"` |
| `ensure` | Fails if value not truthy | `x > 0 \|ensure "must be positive"` |
| `^ensure` | Like ensure, but also returns immediately | `x > 0 \|^ensure "must be positive"` |
| `fix` | Executes block if failed, recovers | `val \|fix { default }` |
| `^fix` | Like fix, but returns immediately | `val \|^fix { default }` |
| `fix\either` | Two blocks: one for error, one for success | `fix\either val { err } { ok }` |
| `fix\else` | Executes block only if not failed | `val \|fix\else { process }` |
| `fix\continue` | Error/success branching with continuation | `val \|fix\continue { err } { ok }` |
| `continue` | Executes block only if not failed | `val \|continue { process }` |
| `^fix\match` | Pattern matches error codes with handlers | `^fix\match err { 404 { "not found" } }` |

### Control Flow

| Function | Description | Example |
|----------|-------------|---------|
| `try` | Executes block, clears flags afterward | `try { risky-op }` |
| `try-all` | Returns `[success result]` tuple | `[ok res]: try-all { op }` |
| `try\in` | Executes block in given context | `try\in ctx { op }` |
| `finally` | Ensures cleanup block runs regardless | `finally { op } { cleanup }` |
| `retry` | Retries block up to N times on failure | `retry 3 { flaky-op }` |
| `persist` | Retries until success (max 1000) | `persist { until-ready }` |
| `timeout` | Fails if block exceeds time limit (ms) | `timeout 5000 { slow-op }` |
| `requires-one-of` | Validates value is in allowed set | `val \|requires-one-of { "a" "b" }` |
| `^requires-one-of` | Like above, but returns immediately | `val \|^requires-one-of { 1 2 }` |

---

---

# Blogpost brainstorm: “Try-catch is still the GOTO”

This section is an outline / idea bank for a blogpost (can be extracted into its own file later).

## Thesis

`try/catch` is still at the **GOTO level** of abstraction:

- **Unstructured jump**: evaluation jumps from an *unknown* point inside the try-block to a *distant* handler block.
- **Low intent**: `catch` doesn’t say *what kind* of failure you intend to handle. It’s a “fix-all” mechanism.
- **Overpowered**: can emulate branching, early returns, loop control, even non-local exits — like `goto`.

## Levels of intent (control flow)

1. **GOTO-level**: unstructured jumps (goto, exceptions as a non-local jump).
2. **Structured control**: `if/else`, `while/until`, `for`.
3. **Higher-order intent**: `map`, `filter`, `reduce`, `group-by`, `pipeline`.

The claim: mainstream languages evolved from (1) → (2) → (3) for “success paths”, but error paths often stayed at (1).

## How Rye pushes failures up the intent ladder

- `check` = add *context* (error-chaining) — intentional enrichment.
- `ensure` / `requires-one-of` = intentional *validation*.
- `fix` / `fix\either` / `fix\continue` = intentional *recovery / branching*.
- `finally` / `defer` = intentional *cleanup*.
- `retry` / `timeout` / `persist` = intentional *operational policy*.

So instead of `catch (Exception e)` you express: “add context here”, “recover here”, “validate here”, “retry here”.

## Suggested structure

1. **Open with the accusation**: try/catch is a goto wearing a suit.
2. **Why the jump is worse than goto**: the jump originates from *unknown depth* in a block; you lose local reasoning.
3. **What we did for success paths**: structured programming; then higher order functions.
4. **Error handling lagged behind**: exceptions stayed unstructured.
5. **Rye’s model**: failure as value + explicit flag, and combinators as higher-intent building blocks.
6. **Examples**: show a pipeline with `check`/`ensure`/`fix` vs nested try/catch.
7. **Closing**: “don’t abolish exceptions; lift error handling to higher intent”.

## Soundbites

- “Catch has no intent; it’s a fixall.”
- “Exceptions are control flow without syntax.”
- “If your error-handling primitive can model loops, branches, and early returns, it’s probably a goto.”
