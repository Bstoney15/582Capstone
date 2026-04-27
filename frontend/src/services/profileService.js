const API_BASE = "/api/user/profile";

async function readError(response, fallbackMessage) {
  const text = await response.text();
  return text || fallbackMessage;
}

export async function fetchUserProfile() {
  const response = await fetch(API_BASE, {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error(await readError(response, "Failed to fetch profile"));
  }

  return response.json();
}

export async function updateUserProfile(payload) {
  const response = await fetch(API_BASE, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    throw new Error(await readError(response, "Failed to update profile"));
  }

  return response.json();
}