// Login form handler for Mitto auth page

// Get API prefix (injected by server into the page)
function getApiPrefix() {
  return window.mittoApiPrefix || "";
}

document.addEventListener("DOMContentLoaded", function () {
  const form = document.getElementById("loginForm");
  const errorDiv = document.getElementById("error");
  const submitBtn = document.getElementById("submitBtn");
  const cloudflareMsg = document.getElementById("cloudflare-message");

  if (!form || !errorDiv || !submitBtn) {
    console.error("Required form elements not found");
    return;
  }

  // Fetch auth info to adapt the UI before showing the form
  fetch(getApiPrefix() + "/api/auth-info")
    .then(function (res) {
      return res.json();
    })
    .then(function (info) {
      if (!info.simple && info.cloudflare) {
        // Only Cloudflare auth configured: hide login form, show message
        form.style.display = "none";
        if (cloudflareMsg) {
          cloudflareMsg.style.display = "";
        }
      } else if (!info.simple && !info.cloudflare) {
        // No auth configured: show a generic notice in the error div
        errorDiv.textContent = "No authentication method is configured.";
        errorDiv.classList.remove("hidden");
        form.style.display = "none";
      }
      // If info.simple is true, show the normal login form (default state)
    })
    .catch(function (err) {
      console.warn("Failed to fetch auth info:", err);
      // On error, keep the login form visible so the user can still try
    });

  form.addEventListener("submit", async function (e) {
    e.preventDefault();

    // Disable form during submission
    submitBtn.disabled = true;
    submitBtn.textContent = "Signing in...";
    errorDiv.classList.add("hidden");

    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    try {
      const response = await fetch(getApiPrefix() + "/api/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username: username, password: password }),
        credentials: "same-origin", // Include cookies in request and accept Set-Cookie
      });

      if (response.ok) {
        // Redirect to main app
        window.location.href = "/";
      } else {
        const data = await response.json().catch(function () {
          return {};
        });
        errorDiv.textContent = data.error || "Invalid username or password";
        errorDiv.classList.remove("hidden");
      }
    } catch (err) {
      errorDiv.textContent = "Network error. Please try again.";
      errorDiv.classList.remove("hidden");
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = "Sign In";
    }
  });
});
