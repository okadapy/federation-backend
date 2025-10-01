Эндпоинты:

Метод	Путь	Описание	Параметры	Тело запроса
GET	/user	Получить список всех пользователей	username (опционально)	-
GET	/user/:id	Получить пользователя по ID	id (path)	-
POST	/user	Создать нового пользователя	-	{"username": "string", "password": "string"}
PUT	/user/:id	Обновить пользователя по ID	id (path)	{"username": "string", "password": "string"}
DELETE	/user/:id	Удалить пользователя по ID	id (path)	-
Обратные вызовы (CallBack)
Модель:

json
{
"name": "string",
"phone": "string",
"email": "string",
"team_name": "string",
"callback_type": "string"
}
Эндпоинты:

Метод	Путь	Описание	Параметры	Тело запроса
GET	/callback	Получить список всех обратных вызовов	name, phone, email, team_name, callback_type (опционально)	-
GET	/callback/:id	Получить обратный вызов по ID	id (path)	-
POST	/callback	Создать новый обратный вызов	-	{"name": "string", "phone": "string", "email": "string", "team_name": "string", "callback_type": "string"}
PUT	/callback/:id	Обновить обратный вызов по ID	id (path)	{"name": "string", "phone": "string", "email": "string", "team_name": "string", "callback_type": "string"}
DELETE	/callback/:id	Удалить обратный вызов по ID	id (path)	-
Галерея (GalleryItem)
Модель:

json
{
"images": ["array of file IDs"],
"chapter_id": "number"
}
Эндпоинты:

Метод	Путь	Описание	Параметры	Тело запроса
GET	/gallery	Получить все элементы галереи	-	-
GET	/gallery/:id	Получить элемент галереи по ID	id (path)	-
POST	/gallery	Создать новый элемент галереи	-	{"images": [array of file IDs], "chapter_id": number}
PUT	/gallery/:id	Обновить элемент галереи по ID	id (path)	{"images": [array of file IDs], "chapter_id": number}
DELETE	/gallery/:id	Удалить элемент галереи по ID	id (path)	-
Новости (News)
Модель:

json
{
"heading": "string",
"description": "string",
"images": ["array of file IDs"],
"date": "timestamp",
"chapter_id": "number"
}
Эндпоинты:

Метод	Путь	Описание	Параметры	Тело запроса
GET	/news	Получить список всех новостей	-	-
GET	/news/:id	Получить новость по ID	id (path)	-
POST	/news	Создать новую новость	-	{"heading": "string", "description": "string", "images": [array of file IDs], "date": "timestamp", "chapter_id": number}
PUT	/news/:id	Обновить новость по ID	id (path)	{"heading": "string", "description": "string", "images": [array of file IDs], "date": "timestamp", "chapter_id": number}
DELETE	/news/:id	Удалить новость по ID	id (path)	-
Разделы (Chapter)
Модель:

json
{
"name": "string",
"page": "string"
}
Эндпоинты:

Метод	Путь	Описание	Параметры	Тело запроса
GET	/chapter	Получить список всех разделов	name, page (опционально)	-
GET	/chapter/:id	Получить раздел по ID	id (path)	-
POST	/chapter	Создать новый раздел	-	{"name": "string", "page": "string"}
PUT	/chapter/:id	Обновить раздел по ID	id (path)	{"name": "string", "page": "string"}
DELETE	/chapter/:id	Удалить раздел по ID	id (path)	-
Команды (Team)
Модель:

json
{
"team_name": "string",
"sex": "string",
"team_logo_id": "number"
}
Эндпоинты:

Метод	Путь	Описание	Параметры	Тело запроса
GET	/team	Получить список всех команд	team_name, sex, team_logo_id (опционально)	-
GET	/team/:id	Получить команду по ID	id (path)	-
POST	/team	Создать новую команду	-	{"team_name": "string", "sex": "string", "team_logo_id": number}
PUT	/team/:id	Обновить команду по ID	id (path)	{"team_name": "string", "sex": "string", "team_logo_id": number}
DELETE	/team/:id	Удалить команду по ID	id (path)	-
Особенности фильтрации
Для эндпоинтов с CRUD контроллерами (/user, /callback, /chapter, /team) доступна фильтрация через query parameters. Можно фильтровать по любому полю модели:

GET /user?username=admin

GET /team?sex=male&team_name=TeamA

GET /callback?callback_type=team_application