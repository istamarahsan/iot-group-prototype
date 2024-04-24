from transformers import pipeline
from fastapi import FastAPI, Request
import os
import http

classifier_pipe = pipeline("audio-classification", model="MIT/ast-finetuned-audioset-10-10-0.4593")

app = FastAPI()

TOKEN = os.environ["ACCESS_KEY"]

@app.get("/health")
async def healthcheck(request: Request):
    return http.HTTPStatus.OK if TOKEN != "" and request.headers['Authorization'] == f"Bearer {TOKEN}" else http.HTTPStatus.UNAUTHORIZED

@app.post("/classify")
async def classify(request: Request):
    if TOKEN != "" and request.headers['Authorization'] == f"Bearer {TOKEN}":
        return http.HTTPStatus.UNAUTHORIZED

    bytes = await request.body()
    result = classifier_pipe(bytes)
    return result