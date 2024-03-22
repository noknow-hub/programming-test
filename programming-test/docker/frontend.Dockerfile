# イメージベース ※実際のプロジェクトでは、おそらくver.を揃える。今回はテストなのでlatestで
FROM httpd:latest

# Docker起動マシンから各種ファイルを仮想環境へコピー
COPY ./frontend /usr/local/apache2/htdocs/