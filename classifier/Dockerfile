FROM python:3.12-bookworm AS base

RUN apt update && apt install ffmpeg -y

FROM base AS builder

ENV PIPENV_VENV_IN_PROJECT=1
RUN pip install pipenv --user

ADD Pipfile.lock Pipfile /usr/src/
WORKDIR /usr/src
RUN /root/.local/bin/pipenv requirements > requirements.txt

FROM base AS run

COPY --from=builder /usr/src/requirements.txt /usr/src

ADD main.py /usr/src/

WORKDIR /usr/src
RUN pip install uvicorn
RUN pip install -r requirements.txt

CMD ["uvicorn", "main:app"]