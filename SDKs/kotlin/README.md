# ai-router Kotlin SDK

Kotlin SDK for the [ai-router](../../) LLM proxy. Build requests with a type-safe DSL, send them to a local or remote proxy, and consume responses — including structured output via `@Serializable` classes.

## Installation

Add the SDK to your `build.gradle.kts`:

```kotlin
dependencies {
    implementation("ai.router:ai-router-sdk:0.1.0")
}
```

> For now, build from source: `./gradlew publishToMavenLocal`

## Quick Start

```kotlin
import ai.router.sdk.AiRouterClient
import ai.router.sdk.dsl.chatRequest
import ai.router.sdk.models.ReasoningEffort

val client = AiRouterClient("http://localhost:8787")

client.use {
    val response = it.chat(chatRequest("llama3.2:local") {
        messages {
            system { text("You are a helpful assistant.") }
            user { text("What is the capital of France?") }
        }
    })
    println(response.textContent)
}
```

## Chat Completion

### Multi-Turn Conversation

```kotlin
val request = chatRequest("llama3.2:local") {
    messages {
        system { text("You are a helpful assistant.") }
        user { text("My name is Alice.") }
        assistant { text("Hello Alice! How can I help you?") }
        user { text("What's my name?") }
    }
    temperature(0.7)
    maxTokens(512)
}
```

### Vision (Multimodal)

```kotlin
val request = chatRequest("llava:local") {
    messages {
        user {
            text("What's in this image?")
            image(mimeType = "image/png", base64Data = "iVBOR...")
        }
    }
}
```

### Tool Calling

```kotlin
val request = chatRequest("llama3.2:cloud") {
    messages {
        user { text("What's the weather in Berlin?") }
    }
    tools {
        tool("get_weather", "Get current weather for a city") {
            put("type", JsonPrimitive("object"))
            put("properties", JsonObject(mapOf(
                "city" to JsonObject(mapOf("type" to JsonPrimitive("string")))
            )))
            put("required", JsonArray(listOf(JsonPrimitive("city"))))
        }
    }
}

// After getting tool calls in the response, send results back:
val followUp = chatRequest("llama3.2:cloud") {
    messages {
        user { text("What's the weather in Berlin?") }
        // re-add the assistant message with tool_calls from response...
        tool(callId = "call_abc123") {
            text("""{"temp_celsius": 22, "condition": "sunny"}""")
        }
    }
}
```

### Structured Output

Define your response shape as a `@Serializable` class and the SDK handles schema generation + response parsing:

```kotlin
@Serializable
data class WeatherInfo(val city: String, val tempCelsius: Double, val condition: String)

val request = chatRequest("gpt-4:cloud") {
    messages {
        system { text("Extract weather information from the text.") }
        user { text("It's 22°C and sunny in Berlin today.") }
    }
    structuredOutput<WeatherInfo>()
}

val weather: WeatherInfo = client.chat(request, WeatherInfo.serializer())
// weather.city == "Berlin", weather.tempCelsius == 22.0, etc.
```

### Reasoning

```kotlin
val request = chatRequest("deepseek-r1:cloud") {
    messages {
        user { text("Prove that √2 is irrational.") }
    }
    reasoningEffort(ReasoningEffort.HIGH)
}

val response = client.chat(request)
println("Reasoning: ${response.choices.message.reasoningContent}")
println("Answer: ${response.textContent}")
```

## Embeddings

```kotlin
import ai.router.sdk.dsl.embedRequest

val request = embedRequest("nomic-embed-text:local") {
    batch("First document to embed")
    batch("Second document to embed")
    batch("Third document to embed")
    dimensions(768)
}

val response = client.embed(request)
response.data.forEach { embedding ->
    println("Embedding ${embedding.index}: ${embedding.embedding.size} dimensions")
}
```

## Error Handling

```kotlin
try {
    client.chat(request)
} catch (e: AiRouterException) {
    println("Status: ${e.statusCode}")
    println("Type: ${e.apiError.type}")
    println("Message: ${e.apiError.message}")
}
```

## Configuration

```kotlin
// Custom Ktor client for full control over timeouts, logging, etc.
val customClient = HttpClient(CIO) {
    install(ContentNegotiation) { json() }
    install(HttpTimeout) {
        requestTimeoutMillis = 300_000
    }
}

val client = AiRouterClient("http://my-remote-proxy:8787", customClient)
```
