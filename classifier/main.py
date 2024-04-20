from transformers import pipeline
from fastapi import FastAPI, Request

classifier_pipe = pipeline("audio-classification", model="MIT/ast-finetuned-audioset-10-10-0.4593")

app = FastAPI()

@app.get("/health")
async def healthcheck():
    return 200

@app.post("/classify")
async def classify(request: Request):
    bytes = await request.body()
    result = classifier_pipe(bytes)
    return result