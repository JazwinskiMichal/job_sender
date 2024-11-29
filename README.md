# Scheduling and Timesheet Management System

A web application built with Go that manages contractor timesheets and schedules, featuring server-side rendering and Google Cloud Platform integration.

## Features

- User authentication and role-based access control
- Group and contractor management
- Automated email notifications
- Timesheet upload and processing
- Scheduled tasks execution

## Tech Stack

### Backend
- Go
- Firebase/Firestore
- Google Cloud Platform
  - Cloud Run
  - Cloud Tasks
  - Cloud Scheduler
  - Secret Manager
  - Cloud Storage

### Security
- Session-based authentication
- Secure secret management
- Middleware-based panic recovery
- Role-based access control

## Architecture

The application follows a structured architecture with:
- Separate service layers for core functionality
- Middleware for authentication and error handling
- Handler-based routing system
- Environment-based configuration
