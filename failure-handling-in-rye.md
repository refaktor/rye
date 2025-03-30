# Embracing Failure: Error Handling in Rye vs Python

In most programming languages, errors are something to be feared - unwelcome interruptions that break your flow and force you to write defensive code. But in Rye, we take a different approach. Failures are first-class citizens, designed to be created, inspected, transformed, and handled with elegance.

Let's dive into Rye's failure handling system and see how it compares to Python's exception handling, highlighting how Rye's approach can make your code more robust and expressive.

## The Three Faces of Failure

Rye offers three main ways to create errors, each with its own behavior:

```rye
// Creates an error and sets the failure flag, but continues execution
fail "Something went wrong"

// Creates an error, sets the failure flag, AND immediately returns from the function
^fail "Something went catastrophically wrong"

// Just creates an error object without setting any flags
err: failure "This is just an error value"
```

The difference is subtle but powerful. `fail` signals a problem but lets the surrounding code decide what to do, while `^fail` is more decisive, immediately exiting the current function. And `failure` just creates an error value without affecting program flow at all.

### Python Comparison

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

Unlike Rye, Python doesn't distinguish between creating an error and propagating it - when you `raise` an exception, it immediately starts unwinding the stack. There's no built-in concept similar to Rye's `fail` that creates an error but continues execution.

## Anatomy of an Error

Errors in Rye aren't just strings - they're rich objects that can carry structured information:

```rye
// Create an error with a status code
api-error: fail 404

// Create an error with a detailed message
validation-error: fail "Invalid email format"

// Create an error with both code and additional details
db-error: fail { 
  "code" 1001 
  "table" "users" 
  "operation" "insert" 
}
```

You can then inspect these errors using accessor functions:

```rye
api-error |status?    // Returns 404
validation-error |message?  // Returns "Invalid email format"
db-error |details? .table  // Returns "users"
```

### Python Comparison

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

## The Art of Error Handling

Where Rye really shines is in how it lets you handle errors. Let's look at some patterns:

### The Fix Pattern

The `fix` combinator lets you handle errors by providing a recovery block:

```rye
// Try to get a user, or return a default if it fails
get-user 123 |fix { default-user }

// Try to parse a date, or return today if it fails
parse-date user-input |fix { today }
```

#### Python Comparison

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

### The Check Pattern

The `check` combinator lets you transform errors by wrapping them with additional context:

```rye
// Add context to any error that might occur
db/query "SELECT * FROM users" |check "Error querying users database"

// In a chain of operations, add context at each step
file/read "config.json" 
  |check "Failed to read config file" 
  |json/parse 
  |check "Config file contains invalid JSON"
```

#### Python Comparison

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

### The Ensure Pattern

The `ensure` combinator lets you validate conditions and fail if they're not met:

```rye
// Validate user input
user-age > 0 |ensure "Age must be positive"
user-email |contains "@" |ensure "Email must contain @"

// Check preconditions before proceeding
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

## Real-World Examples

Let's look at some practical examples of how these patterns come together:

### Example 1: API Request with Error Handling

```rye
fetch-user-data: fn { user-id } {
  // Validate input
  user-id > 0 |^ensure "User ID must be positive"
  
  // Try to fetch the user, with multiple layers of error handling
  http/get (join "https://api.example.com/users/" user-id)
    |check "API request failed" 
    |json/parse 
    |check "Invalid JSON response"
    |fix { 
      // If anything failed, return a default user
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
  // Retry the database operation up to 3 times
  retry 3 {
    // Ensure we have a valid user
    user/valid? user |^ensure "Invalid user data"
    
    // Try to save, with timeout protection
    timeout 5000 {
      db/begin-transaction
      
      // Use defer to ensure transaction is rolled back on any failure
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
  // A chain of operations where any step can fail
  validate-order order
    |fix\continue {
      // Error handler
      log/error "Order validation failed"
      fail "Invalid order"
    } {
      // Success handler - continues the chain
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

## The Philosophy of Failure

Rye's approach to error handling reflects a deeper philosophy: failures are normal, expected parts of a program's execution. Rather than treating them as exceptional cases to be avoided, we embrace them as values that can be passed around, transformed, and handled just like any other data.

This leads to code that's more robust, more expressive, and often more concise than traditional error handling approaches. Instead of deeply nested try/catch blocks or error codes that must be checked after every operation, Rye lets you express your error handling logic as a natural part of your data flow.

### Key Differences from Python

1. **Errors as Values**: In Rye, errors are just values that can be passed around, while in Python, exceptions are special control flow mechanisms.

2. **Composability**: Rye's error handling combinators compose naturally with its pipe-based syntax, while Python's try/except blocks don't compose well.

3. **Granular Control**: Rye gives you fine-grained control over when to propagate, transform, or recover from errors, while Python's exceptions are more all-or-nothing.

4. **Explicit Error Paths**: In Rye, error handling is explicit in the data flow, making it easier to see how errors are handled, while Python's exception handling can be more implicit and harder to follow.

5. **Built-in Patterns**: Rye provides built-in combinators for common error handling patterns, while Python requires you to implement these patterns yourself.

So next time you're writing Rye code, don't fear failure - embrace it! Your code will be better for it.
