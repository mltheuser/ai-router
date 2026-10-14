import type { ModelFilters, ModelList } from "../models";
import type { Transport } from "../transport";
import type { ChatModel, ChatRequest, ChatResponse } from "./chat";

const USE_CASE = "chat";

/** The chat use case: a conversation in, the model's next message out. */
export class ChatClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  send(request: ChatRequest): Promise<ChatResponse> {
    return this.#transport.send(USE_CASE, request);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters): Promise<ModelList<ChatModel>> {
    return this.#transport.listModels(USE_CASE, filters);
  }
}
