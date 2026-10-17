import { expect, test } from "bun:test";
import { wavBase64 } from "ai-router-sdk";
import { file } from "bun";
import { newExampleClient, SPEECH_TO_TEXT_MODEL } from "./setup";

test("Speech to text: transcribe a WAV file", async () => {
  // A WAV file of 16-bit mono PCM at 24 kHz, made by our text to speech. It
  // says "The quick brown fox jumps over the lazy dog."
  const wav = await file(new URL("resources/sentence.wav", import.meta.url)).bytes();
  const client = newExampleClient();

  const response = await client.speechToText.send({
    model: SPEECH_TO_TEXT_MODEL,
    audio: wavBase64(wav),
  });

  console.log(`Transcript: ${response.text}`);
  expect(response.text.trim()).not.toBeEmpty();
});
