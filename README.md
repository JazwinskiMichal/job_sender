# Job Sender

A web application built with Go that provides a comprehensive solution for managing contractor timesheets and scheduling automated email tasks, leveraging Google Cloud Platform services for reliable and secure operation.

## Features

- Server-side rendered web interface
- User authentication and session management
- Automated email scheduling and sending
- Secure file storage for timesheets
- Role-based access control (Owner/Contractor)

## Tech Stack

### Core
- Go 1.22
- Gorilla Web Toolkit (Mux, Sessions)
- Firebase Admin SDK

### Google Cloud Platform Services
- Cloud Run (containerized deployment)
- Cloud Firestore (data storage)
- Cloud Tasks (job queuing)
- Cloud Scheduler (task scheduling)
- Cloud Storage (file storage)
- Secret Manager (secure configuration)
- Error Reporting

### Email Integration
- IMAP support via go-imap
- SMTP email sending
- Email template system

## Key Components

### Authentication & Security
- Firebase Authentication
- Secure session management
- Secret Manager integration for credentials
- Middleware-based authorization

### Task Management
- Automated email scheduling
- Task queue management
- File upload and processing
- Error reporting and monitoring

### Data Management
- Contractor information
- Group management
- Timesheet processing
- Owner administration

## API Endpoints

### Authentication
- `GET /login` - Show login page
- `POST /login` - Authenticate user
- `POST /logout` - Log out user

### Main
- `GET /main` - Main application entry point
- `GET /` - Redirects to main

### Owners
- `GET /auth/owners/add` - Show add owner form
- `GET /auth/owners/{ID}` - Get owner details
- `GET /auth/owners/{ID}/edit` - Show edit owner form
- `POST /auth/owners` - Create new owner
- `PUT /auth/owners/{ID}` - Update owner
- `DELETE /auth/owners/{ID}` - Delete owner

### Groups
- `GET /auth/groups/add` - Show add group form
- `GET /auth/groups/{ID}` - Get group details
- `GET /auth/groups/{ID}/edit` - Show edit group form
- `GET /auth/groups/{ID}/delete` - Delete group
- `POST /auth/groups` - Create new group
- `POST /auth/groups/{ID}` - Update group

### Contractors
- `GET /auth/contractors` - Get contractors for a group
- `GET /auth/contractors/add` - Show add contractor form
- `GET /auth/contractors/{ID}/edit` - Show edit contractor form
- `POST /auth/contractors` - Add new contractor
- `POST /auth/contractors/{ID}` - Update contractor
- `DELETE /auth/contractors/{ID}` - Delete contractor

### Registration & Authentication
- `GET /login` - Show login page
- `POST /login` - Authenticate user
- `POST /logout` - Log out user
- `GET /register` - Show registration form
- `GET /register/confirm` - Show registration confirmation
- `POST /register` - Create new user account

### Timesheets
- `POST /timesheets/request` - Send timesheet request to contractors
- `POST /timesheets/aggregate` - Process and store timesheet submissions
  - Handles email attachments
  - Stores files in Cloud Storage
  - Updates contractor records
  - Archives processed emails

### Error Handling
- `GET /somethingWentWrong` - Display error page for system errors

## Security

- Session-based authentication
- Secure secret management
- Middleware-based panic recovery
- Role-based access control