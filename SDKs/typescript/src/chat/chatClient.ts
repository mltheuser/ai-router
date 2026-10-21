import type { ModelFilters, ModelList } from "../models";
import type { CallOptions, Transport } from "../transport";
import type { ChatMessage } from "./chatMessage";
import type { ChatModel } from "./chatModel";
import { chatRequest } from "./chatRequest";
import type { ChatResponse } from "./chatResponse";
import type { ChatSettings } from "./chatSettings";

const USE_CASE = "chat";

/** The chat use case: a conversation in, the model's next message out. */
export class ChatClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  /** The model's next message in the conversation `messages`, answered as `settings` say. */
  send(
    settings: ChatSettings,
    messages: ChatMessage[],
    options?: CallOptions,
  ): Promise<ChatResponse> {
    return this.#transport.send(USE_CASE, chatRequest(settings, messages), options);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters, options?: CallOptions): Promise<ModelList<ChatModel>> {
    return this.#transport.listModels(USE_CASE, filters, options);
  }
}
