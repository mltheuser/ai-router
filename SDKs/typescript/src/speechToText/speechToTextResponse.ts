/** The answer to a speech-to-text request: `text` is the transcript of the whole audio. */
export interface SpeechToTextResponse {
  model: string;
  text: string;
}
