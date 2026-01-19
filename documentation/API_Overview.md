# Document and Share API Overview

This document explains the core flows for managing documents and sharing them with other users.

## 1. Create Document
**Purpose**: Creates a new document. If a title is not provided, it defaults to "Untitled Document".
- **Endpoint**: `POST /v1/documents`
- **Request Body**:
  ```json
  {
    "title": "My New Project" 
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": "uuid-string",
      "title": "My New Project",
      "content": "",
      "role": "owner",
      "created_at": "...",
      "updated_at": "..."
    }
  }
  ```

## 2. Rename or Update Document
**Purpose**: Update the document's title (rename) or content.
- **Endpoint**: `PATCH /v1/documents/:id`
- **Request Body**:
  ```json
  {
    "title": "Renamed Title",
    "content": "Updated content..."
  }
  ```
- **Response**: Returns the updated document object.

## 3. Share Document (Invite Users)
**Purpose**: Grant access to specific users via their email. You can assign them `viewer` or `editor` roles.
- **Endpoint**: `POST /v1/documents/:id/share`
- **Request Body** (Array of invites):
  ```json
  [
    {
      "email": "colleague@example.com",
      "role": "editor"
    },
    {
      "email": "friend@example.com",
      "role": "viewer"
    }
  ]
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "Invitations processed",
    "data": {
      "invited": ["colleague@example.com", ...]
    }
  }
  ```

## 4. List Permissions
**Purpose**: View all users who have access to this document and their respective roles. Useful for displaying a "Manage Access" modal.
- **Endpoint**: `GET /v1/documents/:id/share`
- **Response**:
  ```json
  {
    "success": true,
    "data": [
      {
        "user_id": "uuid-1",
        "email": "owner@example.com",
        "name": "Owner Name",
        "role": "owner"
      },
      {
        "user_id": "uuid-2",
        "email": "colleague@example.com",
        "name": "Colleague Name",
        "role": "editor"
      }
    ]
  }
  ```
## 5. Request Access
**Purpose**: Request access to a document the user currently doesn't have permission to view.
- **Endpoint**: `POST /v1/documents/:id/request-access`
- **Request Body**:
  ```json
  {
    "role": "editor" // or "viewer"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "Access requested",
    "data": null
  }
  ```

## 6. List Access Requests
**Purpose**: View all pending access requests. Returns requests sent by the user and requests received by the user (for documents they own).
- **Endpoint**: `GET /v1/access/requests`
- **Response**:
  ```json
  {
    "success": true,
    "message": "Requests listed",
    "data": {
      "sent": [
        {
          "id": "req-uuid-1",
          "document_id": "doc-uuid",
          "document_title": "Project Alpha",
          "status": "pending",
          "created_at": "..."
        }
      ],
      "received": [
        {
          "id": "req-uuid-2",
          "document_id": "doc-uuid",
          "document_title": "My Secret Doc",
          "requester_email": "stranger@example.com",
          "status": "pending",
          "created_at": "..."
        }
      ]
    }
  }
  ```

## 7. Handle Access Request
**Purpose**: Approve or reject an access request.
- **Endpoint**: `POST /v1/documents/requests/:requestId`
- **Request Body**:
  ```json
  {
    "action": "approve", // or "reject"
    "role": "editor"     // required if action is "approve"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "Request handled",
    "data": null
  }
  ```
