# Synlabs project

## How to run
Project uses postgres database. Create a postgres server and update its details in the .env file.
Create a database with name 'synlabs'

Run the project using 
```
go run .
```
Project is hosted on localhost:8080

Open https://localhost:8080/ping to check availability

## Sample Screenshots

### POST /signup
<img width="972" alt="image" src="https://github.com/user-attachments/assets/b71b7494-b8f9-4e7d-851b-1b2821d5ca72">

### POST /login
<img width="971" alt="image" src="https://github.com/user-attachments/assets/1abccadd-cb57-4551-ba7d-4dbd187725c5">

### POST /uploadResume
<img width="964" alt="image" src="https://github.com/user-attachments/assets/dc1a4b98-6de5-4fd0-a7af-860cdd155f9b">

### POST /admin/job
<img width="920" alt="image" src="https://github.com/user-attachments/assets/cd7b6b45-b21a-41ef-9a3e-b65f8511ded5">

### GET /admin/job/{job_id}
<img width="965" alt="image" src="https://github.com/user-attachments/assets/21bacbd5-8806-40c6-8241-4fc7cd521347">

### GET /admin/applicants
<img width="995" alt="image" src="https://github.com/user-attachments/assets/9ae8fde1-147b-49e3-864f-5f169c6dcd45">

### GET /admin/applicant/{applicant_id}
<img width="990" alt="image" src="https://github.com/user-attachments/assets/00f0dbc8-0c82-4028-9463-107d21848d28">

### GET /jobs
<img width="967" alt="image" src="https://github.com/user-attachments/assets/eb0dd2ef-d234-433d-8cf5-63b867c19c1b">

### GET /jobs/apply?job_id={job_id}
<img width="909" alt="image" src="https://github.com/user-attachments/assets/81704ac5-aa31-41b7-8162-e6417f74787a">

