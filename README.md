# ⚡ OrchAxon AutoTest CLI

> **AutoTest Gen** is a developer-first tool to generate robust Unit Test boilerplate instantly.
> Supports **C#, Kotlin, TypeScript, Java, PHP, and Node.js**.

Developed by **[OrchAxon Labs](https://github.com/Mr-Fullstack)**.

---

## 🚀 Features

- **Zero Config:** Just run `autotest -lang node -class User` and get the code.
- **Meta-Framework Support:** Define complex scenarios with Mocks using a simple JSON file.
- **Multi-Language:** Supports syntax and mocking libraries for 8 languages.
- **Blazing Fast:** Written in Go, compiles to a single static binary.
- **Smart Output:** Automatically creates a `test/` folder and saves the file with the correct extension.
- **Batch Processing:** Generate tests for hundreds of specs at once using wildcards.
- **Multi-Language Specs:** Define one spec, generate tests for multiple languages simultaneously.
- **Auto-Organization:** Automatically creates a test/ folder and names files based on conventions (.test.js, Test.cs, _test.go).
---

## 📦 Installation

### Option 1: Via Go (Recommended)
If you have Go installed, this is the fastest way:
```bash 
go install github.com/Mr-Fullstack/orchaxon-autotest@latest
```

### Option 2: Manual Download (Binaries)
#### You don't need Go installed. Just download the pre-compiled binary for your OS:
1. Go to the [Download](https://github.com/Mr-Fullstack/orchaxon-autotest/releases) 
2. Download the .exe (Windows), linux or mac binary
3. Add it to your PATH (optional) or run it directly.

### 🛠 Usage 
#### 1.Simple Mode (Quick Boilerplate)
Generate a test file automatically in 3. the test/ folder.

**Syntax:**
```bash  
autotest -lang <language> -class <ClassName>
```

**Examples:**
```bash 
# Generate C# xUnit Test
autotest -lang csharp -class OrderService

# Generate Node.js Native Test
autotest -lang node -class AuthService

# Generate Kotlin JUnit Test
autotest -lang kotlin -class PaymentProcessor
```

### Flag Dictionary (lang):
| Lang                  | file    | 
|-----------------------|---------|
| php                   | .php    |
| java                  | .java   |
| kt, kotlin            | .kt     |
| go, golang            | .go     |
| cs, csharp            | .cs     |
| ts, typescript        | .ts     |
| js, node, javascript  | .js     |

#### 2. Advanced Mode (JSON Spec)

 For complex scenarios with Mocks, Dependency Injection, and specific test cases, 
create a spec.json file.1. 

**1. Create a file named auth_spec.json:**
```json 
{
  "meta": { 
    "langs": ["csharp", "node", "go"], // 🔥 Generates 3 files at once!
    //"lang": "node" // Generate 1 file same!
  },
  "target": {
    "class_name": "AuthService",
    "method_name": "login"
  },
  "dependencies": [
    { "field_name": "db", "interface_name": "Database" },
    { "field_name": "mailer", "interface_name": "EmailService" }
  ],
  "scenarios": [
    {
      "id": "should_return_token",
      "description": "Should return JWT token on success",
      "mocks_setup": [
        { "dependency": "db", "method": "findUser", "return_value": "{ id: 1 }" }
      ],
      "expectations": { "return_value": "'jwt_token'" }
    }
  ]
}
```

**2. Run the tool pointing to the file:**
```bash  
autotest -file auth_spec.json
```


#### 3. Batch Mode (Mass Generation)

Have a folder full of specs? Process them all in a single command using wildcards.

```bash  
autotest -file "specs/*.json"
```

### Supported Languages: 
| Language  | Framework | 
|-----------|-----------|
| Node.js   | Native Runner (node:test)|
| C#        | xUnit + NSubstitute |
| Kotlin    | JUnit + MockK|
| Java      | JUnit 5 + Mockito |
| TypeScript| Jest   |
| PHP       | PHPUnit|
| Go(Golang)| Testify (assert/mock)|

### License
#### MIT  ©[OrchAxon Labs]()