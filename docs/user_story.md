# MailForge User Experience & Use Cases

## Overview

MailForge is an email campaign platform designed for individuals and small businesses that want to send bulk emails easily.

Target users include:

* Individuals sending personal event invitations or reminders
* Small business owners promoting products or services
* Freelancers and service providers communicating with customers

The platform helps users:

* Create an account
* Manage subscribers and email lists
* Create campaigns
* Send emails instantly or schedule them
* Track delivery and engagement performance

---

# User Experience 1: Personal Event Sender

## Persona

**Name:** Sarah
**Age:** 29
**Occupation:** Event Planner
**Use Case:** Sending wedding invitation reminders to friends and family

---

## Problem

Sarah is planning her wedding and has over 300 guests.
She wants to send email reminders to all guests about:

* Wedding date
* Venue
* Dress code
* RSVP deadline

Sending individual emails manually would be stressful and time-consuming.

She needs a simple way to send personalized reminder emails to everyone.

---

## User Journey

### Step 1: Account Registration

Sarah visits MailForge and creates an account.

She provides:

* Full name
* Email
* Password

System actions:

* Account created
* Password hashed
* JWT issued
* User redirected to dashboard

---

### Step 2: Create Subscriber List

Sarah creates a list called:

```text
Wedding Guests
```

System stores:

* List name
* User ownership
* Creation timestamp

---

### Step 3: Add Subscribers

Sarah uploads guest emails manually or through CSV.

Example subscribers:

| Name       | Email                                     |
| ---------- | ----------------------------------------- |
| John Doe   | [john@gmail.com](mailto:john@gmail.com)   |
| Mary Jane  | [mary@gmail.com](mailto:mary@gmail.com)   |
| James Bond | [james@gmail.com](mailto:james@gmail.com) |

System stores:

* Subscriber info
* List association

---

### Step 4: Create Campaign

Sarah creates a campaign.

Campaign details:

```text
Campaign Name: Wedding Reminder
Subject: Reminder - Sarah & Michael Wedding Ceremony
```

Email body:

```text
Dear Guest,

We are excited to remind you about our wedding ceremony happening on July 25th.

Venue: Grand Palace Hall  
Time: 10:00 AM  
Dress Code: Formal

We look forward to celebrating with you.

Best regards,  
Sarah & Michael
```

System stores campaign as:

```text
Status: Draft
```

---

### Step 5: Send Campaign

Sarah selects:

* Recipient list = Wedding Guests
* Send now

System actions:

* Campaign status changes to queued
* Send job added to Redis queue
* Sending worker picks up job

---

### Step 6: Email Processing

Worker process:

1. Load campaign
2. Fetch subscriber list
3. Send emails in batches
4. Use provider (SMTP / Resend)
5. Save delivery results

Delivery statuses:

* Sent
* Failed
* Delivered

---

### Step 7: Tracking & Analytics

Sarah views campaign analytics.

Dashboard shows:

* Total recipients
* Sent emails
* Delivered emails
* Open rate

Example:

```text
Total Recipients: 300
Sent: 300
Delivered: 293
Failed: 7
Opened: 210
```

---

## Modules Involved

* auth
* user
* subscribers
* list
* campaign
* sending
* tracking
* analytics

---

## API Flow

```text
Register → Create List → Add Subscribers → Create Campaign → Send → Track Results
```

---

# User Experience 2: Small Business Owner

## Persona

**Name:** Grace
**Age:** 37
**Occupation:** Tailor / Fashion Designer
**Use Case:** Sending promotional emails to customers

---

## Problem

Grace owns a tailoring business.

She frequently launches:

* New clothes
* Discounts
* Seasonal collections

She wants to notify her existing customers via email whenever new products are available.

She needs:

* Easy customer list management
* Campaign scheduling
* Delivery analytics

---

## User Journey

### Step 1: Login

Grace logs into MailForge.

System validates:

* Email
* Password
* JWT session

---

### Step 2: Manage Customer List

Grace already has a list called:

```text
VIP Customers
```

Current subscribers:

```text
1,250 customers
```

She adds 50 new customers from recent sales.

---

### Step 3: Create Promotion Campaign

Grace creates campaign:

```text
Campaign Name: Summer Collection Launch
Subject: New Collection Just Dropped 🔥
```

Email content:

```text
Hello Valued Customer,

Our latest summer collection is now available.

Enjoy premium designs crafted specially for you.

Visit our store or reply to this email to place an order.

Thank you for always choosing us.
```

Campaign saved as draft.

---

### Step 4: Schedule Campaign

Grace decides to send campaign tomorrow at 8:00 AM.

System actions:

* Campaign status = scheduled
* Scheduled send job created
* Job stored in queue

---

### Step 5: Scheduled Sending

At scheduled time:

Worker process:

1. Fetch campaign
2. Load recipients
3. Send emails in batches
4. Retry failures
5. Store results

Provider options:

* SMTP
* Resend

---

### Step 6: Monitor Campaign Performance

Grace checks analytics after sending.

Dashboard shows:

```text
Recipients: 1250
Sent: 1250
Delivered: 1231
Opened: 760
Clicked: 140
Failed: 19
Open Rate: 60.8%
```

---

### Step 7: Business Insight

Grace uses analytics to answer:

* Did customers open the email?
* Which campaign performed best?
* When should future emails be sent?

This helps improve future marketing performance.

---

## Modules Involved

* auth
* user
* subscribers
* list
* campaign
* sending
* tracking
* analytics

---

## API Flow

```text
Login → Manage Subscribers → Create Campaign → Schedule Send → Track Performance
```

---

# Core Product Value

MailForge helps users:

* Manage email contacts
* Organize contacts into lists
* Create campaigns
* Send emails at scale
* Track delivery and engagement
* Improve communication efficiency

---

# Product Vision

MailForge aims to become a lightweight and powerful email campaign platform for:

* Individuals
* Creators
* Freelancers
* Small businesses

The goal is to make email campaign management simple, affordable, and scalable.
