import { ChatClient } from "./chat/chatClient";
import { ContentsClient } from "./contents/contentsClient";
import { EmbeddingClient } from "./embedding/embeddingClient";
import { SearchClient } from "./search/searchClient";
import { SpeechToTextClient } from "./speechToText/speechToTextClient";
import { TextToSpeechClient } from "./textToSpeech/textToSpeechClient";
import { Transport } from "./transport";

/**
 * Client for the ai-router proxy, with one property per use case.
 *
 * A non-2xx response rejects with an `AiRouterError`; a failure of `fetch`
 * itself (e.g. the server is unreachable, or the call was cancelled or timed
 * out) rejects with that failure as is.
 */
export class AiRouterClient {
  readonly chat: ChatClient;
  readonly embedding: EmbeddingClient;
  readonly search: SearchClient;
  readonly contents: ContentsClient;
  readonly textToSpeech: TextToSpeechClient;
  readonly speechToText: SpeechToTextClient;

  /** @param baseUrl Root URL of the ai-router server, e.g. `http://localhost:8787`. */
  constructor(baseUrl: string) {
    const transport = new Transport(baseUrl);
    this.chat = new ChatClient(transport);
    this.embedding = new EmbeddingClient(transport);
    this.search = new SearchClient(transport);
    this.contents = new ContentsClient(transport);
    this.textToSpeech = new TextToSpeechClient(transport);
    this.speechToText = new SpeechToTextClient(transport);
  }
}
