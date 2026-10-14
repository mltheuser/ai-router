import { expect, test } from "bun:test";
import { wavBytes } from "ai-router-sdk";
import { newExampleClient, TEXT_TO_SPEECH_MODEL } from "./setup";

test("Text to speech: speak a sentence in one of the model's voices", async () => {
  const client = newExampleClient();

  // The model's entry lists the voices it speaks in.
  const { data } = await client.textToSpeech.listModels();
  const voice = data.find((entry) => entry.model === TEXT_TO_SPEECH_MODEL)?.voices[0];
  if (voice === undefined) {
    throw new Error(`expected ${TEXT_TO_SPEECH_MODEL} to be listed with a voice`);
  }

  const response = await client.textToSpeech.send({
    model: TEXT_TO_SPEECH_MODEL,
    voice: voice.id,
    text: "Hello from the ai-router TypeScript SDK.",
  });

  const wav = wavBytes(response);
  console.log(`${voice.name}: ${wav.length} bytes of WAV audio`);
  expect(new TextDecoder().decode(wav.subarray(0, 4))).toBe("RIFF");
});
