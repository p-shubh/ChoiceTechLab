Here's a `README.md` file for your project:

```markdown
# Excel to MySQL and Redis

## Description
This project is a Golang-based application that imports data from an Excel file, stores it into a MySQL database, and caches the data in Redis. It provides a simple CRUD (Create, Read, Update, Delete) system to view, edit, and update records both in the database and cache.

## Prerequisites
- Docker
- Docker Compose
- Golang 1.16+
- MySQL
- Redis

## Setup

### Environment Variables
Create a `.env` file in the root directory with the following content:

```env
MYSQL_DBNAME="my_database"
MYSQL_USER="my_user"
MYSQL_PASSWORD="my_password"
MYSQL_HOST="localhost"
MYSQL_PORT="3306"

PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
GOSU_VERSION="1.17"
MYSQL_MAJOR="8.4"
MYSQL_VERSION="8.4.0-1.el9"
MYSQL_SHELL_VERSION="8.4.0-1.el9"


REDIS_ADDR="0.0.0.0:6379"
REDIS_PASSWORD="YOUR_REDIS"


RediGateway="172.17.0.1"
RedisIPAddress="172.17.0.3"
```

### Running with Docker

1. **Build and run the containers:**
   ```sh
   docker-compose up --build
   ```

This command will build the Golang application and run it alongside MySQL and Redis in Docker containers. The application will be accessible at `http://localhost:8080`.

### Building and Running Locally

1. **Install Dependencies:**
   ```sh
   go mod download
   ```

2. **Build the Application:**
   ```sh
   go build -o main .
   ```

3. **Run the Application:**
   ```sh
   ./main
   ```

### Database Setup

Ensure your MySQL database is running and create a table with the following structure:
```sql
CREATE TABLE IF NOT EXISTS records (
	id INT AUTO_INCREMENT,
	first_name VARCHAR(255) NOT NULL,
	last_name VARCHAR(255) NOT NULL,
	company_name VARCHAR(255) NOT NULL,
	address VARCHAR(255) NOT NULL,
	city VARCHAR(255) NOT NULL,
	county VARCHAR(255) NOT NULL,
	postal VARCHAR(255) NOT NULL,
	phone VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL,
	web VARCHAR(255) NOT NULL,
	PRIMARY KEY (id)
);
```

## API Endpoints

- `POST /upload` - Upload an Excel file to import data
- `GET /records` - Get all records
- `PUT /records/:id` - Update a record by ID
- `DELETE /records/:id` - Delete a record by ID

## Usage

### Upload Excel File
Use an API development tool like Postman to upload an Excel file to the `/upload` endpoint. The file should have the following structure:

| first_name | last_name   | company_name          | address      | city             | county | postal   | phone        | email                  | web                             |
|------------|-------------|-----------------------|--------------|------------------|--------|----------|--------------|------------------------|---------------------------------|
| Aleshia    | Tomkiewicz  | Alan D Rosenburg Cpa Pc | 14 Taylor St | St. Stephens Ward | Kent   | CT2 7PP  | 01944-369967 | atomkiewicz@hotmail.com | http://www.alandrosenburgcpapc.co.uk |

### View Records
Access the `/records` endpoint to view the imported data. If the data is not found in Redis, it will be fetched from the MySQL database and cached in Redis for future requests.

### Edit Record
Send a `PUT` request to `/records/:id` with the updated record details to update a specific record. The record will be updated in both the MySQL database and Redis cache.

### Delete Record
Send a `DELETE` request to `/records/:id` to delete a specific record. The record will be removed from both the MySQL database and Redis cache.

## Error Handling
The application includes error handling mechanisms to gracefully handle failures during file upload, data parsing, database operations, etc. Ensure to check the API responses for any error messages and handle them accordingly.

## Optimization and Scalability
The application is designed to handle increased traffic and data volume efficiently. Database queries and Redis operations are optimized for better performance. The architecture can be scaled horizontally to meet the demand.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
```