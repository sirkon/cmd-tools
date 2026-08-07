# Error processing rules

- An error must not be omitted
- You either return an error or log it, not both.
- If a function has two or more places where it returns an error (no matter the nature of this error), then every
  "outer" error gotten from some call must be annotated before a return. Except errors gotten from recursive calls 
  that must not be annotated.
- Annotation text must describe WHAT caused an error, not stating "this is an error". Example: 
  - Right:
    ```go
    file, err := os.Open(name)
    if err != nil {
        return fmt.Errorf("open file: %w", err)
    }
    ```
  - Wrong:
    ```go
    file, err := os.Open(name)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    ```
- The logging is the opposite: you must say "failed to do something" when logging an error. This is because
  it creates the right topology of an error decomposition:
  1. An error description (text)
  2. Annotation chain N (context)
  3. Annotation chain N - 1 (context)
  4. ...
  5. Annotation chain 1 (context)
  6. Root error (the reason)
  