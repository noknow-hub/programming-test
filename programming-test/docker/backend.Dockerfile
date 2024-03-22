# イメージベース ※実際のプロジェクトでは、おそらくver.を揃える。今回はテストなのでlatestで
FROM golang:latest

WORKDIR /app

COPY ./backend/ .

CMD ["go", "run", "backend.go"]