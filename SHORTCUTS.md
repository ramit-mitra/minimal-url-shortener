# Apple iPhone Shortcut for URL Shortener

This document outlines how to create an Apple iPhone Shortcut to interact with your `url-shortener-golang` application, accessible at `https://your-domain.example.com`.

## Prerequisites

Ensure your `url-shortener-golang` application is running and publicly accessible at `https://your-domain.example.com`.

## Creating the Apple Shortcut

Follow these steps in the "Shortcuts" app on your iPhone:

1.  **Open the Shortcuts app** and tap the `+` icon to create a new shortcut.
2.  **Name your shortcut** (e.g., "Create Shortlink").

Now, add the following actions:

### Action 1: Get Input

*   Search for "Text" and add the "Text" action. This allows you to manually paste a URL.
*   **Alternatively (for Share Sheet integration):** Search for "Receive" and use the "Receive" action, ensuring "URLs" is selected as the input type. This lets you share a URL from other apps directly to the shortcut.
*   **Alternatively (for Clipboard):** Search for "Get Clipboard" and add that action. This will take the URL directly from your clipboard.

### Action 2: URL

*   Search for "URL" and add the "URL" action.
*   In the URL field, enter the endpoint of your shortener: `https://your-domain.example.com/`.

### Action 3: Get Contents of URL

*   Search for "Get Contents of URL" and add it.
*   Change the **Method** to `POST`.
*   Tap on "Headers" and add the following headers:
    *   Key: `Content-Type`, Value: `application/json`
    *   Key: `X-API-Key`, Value: `(your API_KEY value)`
*   Tap on "Request Body" and select `JSON`.
*   Add a new field:
    *   Key: `url`
    *   Type: `Text`
    *   Value: Tap on the "Text" field and select the magic variable for the input from Action 1 (e.g., "Shortcut Input" if using "Receive", or "Clipboard" if using "Get Clipboard").
*   (Optional) If you want single-use links or expiry, add these fields:
    *   **For single-use:** Add a field with Key: `single`, Type: `Boolean`, Value: `True` or `False`.
    *   **For expiry:** Add a field with Key: `expires`, Type: `Number`, Value: `(Unix Timestamp)`.

### Action 4: Get Dictionary Value

*   Search for "Get Dictionary Value" and add it.
*   For "Key", type `shortcode`.
*   For "Dictionary", select "Contents of URL" (this refers to the output from the previous action).

### Action 5: Text

*   Search for "Text" and add the "Text" action.
*   In the text field, construct your full short URL: `https://your-domain.example.com/` followed by the "Shortcode" (the output from Action 4). Tap on the magic variable icon (looks like a square with a dashed outline) and select the "shortcode" output from the previous step.
*   The final text in this action will look something like: `https://your-domain.example.com/SHORTCODE_VARIABLE`

### Action 6: Copy to Clipboard

*   Search for "Copy to Clipboard" and add it.
*   The input to this action should automatically be the "Text" from the previous action (your full short URL).

### Action 7 (Optional): Show Notification

*   Search for "Show Notification" and add it.
*   Title: `Shortlink Created!`
*   Body: The "Text" from Action 5 (your full short URL).

## What the Shortcut Does

Once set up, this shortcut will:
1.  Take a URL as input (from clipboard, share sheet, or manual entry).
2.  Send a `POST` request to your `url-shortener-golang` application at `https://your-domain.example.com/` with the long URL.
3.  Extract the generated `shortcode` from the application's JSON response.
4.  Construct the full short URL (e.g., `https://your-domain.example.com/short/YOUR_SHORTCODE`).
5.  Copy this short URL to your iPhone's clipboard.
6.  (Optionally) Display a notification showing the newly created short URL.
