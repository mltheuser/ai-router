package ai.router.sdk

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.client.request.parameter
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.encodeToJsonElement
import kotlinx.serialization.json.jsonPrimitive

/**
 * The two routes every use case is served on, `POST /v1/<use case>` and
 * `GET /v1/<use case>/models`, over one HTTP client.
 */
internal class Transport(private val baseUrl: String, private val http: HttpClient) : AutoCloseable {

    suspend inline fun <reified Req, reified Res> send(useCase: String, request: Req): Res {
        val response = http.post("$baseUrl/v1/$useCase") {
            contentType(ContentType.Application.Json)
            setBody(request)
        }
        return decode(response)
    }

    suspend inline fun <reified M : ModelRef> listModels(
        useCase: String,
        type: ProviderType?,
        search: String?,
    ): ModelList<M> {
        val response = http.get("$baseUrl/v1/$useCase/models") {
            // The query value is the type's wire name, the @SerialName it serializes to.
            type?.let { parameter("type", json.encodeToJsonElement(it).jsonPrimitive.content) }
            search?.let { parameter("search", it) }
        }
        return decode(response)
    }

    /** Decodes a 2xx response into [Res]; any other status throws [AiRouterException]. */
    suspend inline fun <reified Res> decode(response: HttpResponse): Res {
        if (!response.status.isSuccess()) {
            val error = try {
                response.body<ErrorResponse>().error
            } catch (_: Exception) {
                ApiError(type = "unknown", message = response.bodyAsText())
            }
            throw AiRouterException(response.status.value, error)
        }
        return response.body()
    }

    override fun close() {
        http.close()
    }

    companion object {
        val json: Json = Json {
            ignoreUnknownKeys = true
            encodeDefaults = false
            isLenient = true
        }
    }
}
