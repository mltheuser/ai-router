import { ChatClient } from "./chat/chatClient";
import { ContentsClient } from "./contents/contentsClient";
import { EmbeddingClient } from "./embedding/embeddingClient";
import { SearchClient } from "./search/searchClient";
import { TextToSpeechClient } from "./textToSpeech/textToSpeechClient";
import { type Fetch, Transport } from "./transport";

/** The configuration of an {@link AiRouterClient}. */
export interface AiRouterClientOptions {
  /** Root URL of the ai-router server, e.g. `http://localhost:8787`. */
  baseUrl: string;
  /**
   * The HTTP layer every request goes through, e.g. to add timeouts or
   * logging. Defaults to the global `fetch`.
   */
  fetch?: Fetch;
}

/**
 * Client for the ai-router proxy, with one property per use case.
 *
 * A non-2xx response rejects with an `AiRouterError`; a failure of `fetch`
 * itself (e.g. the server is unreachable, or a timeout aborts the request)
 * rejects with that failure as is.
 */
export class AiRouterClient {
  readonly chat: ChatClient;
  readonly embedding: EmbeddingClient;
  readonly search: SearchClient;
  readonly contents: ContentsClient;
  readonly textToSpeech: TextToSpeechClient;

  constructor(options: AiRouterClientOptions) {
    const transport = new Transport(options.baseUrl, options.fetch ?? fetch);
    this.chat = new ChatClient(transport);
    this.embedding = new EmbeddingClient(transport);
    this.search = new SearchClient(transport);
    this.contents = new ContentsClient(transport);
    this.textToSpeech = new TextToSpeechClient(transport);
  }
}
