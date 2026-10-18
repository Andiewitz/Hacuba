import { expect, test, type Page } from "@playwright/test";

const password = "E2ePassword123";

function uniqueEmail(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2)}@hacuba.test`;
}

async function register(page: Page, email: string) {
  await page.goto("/");
  await page.getByRole("button", { name: "Account menu" }).click();
  await page.getByRole("menuitem", { name: "Log in / Sign up" }).click();

  const dialog = page.getByRole("dialog");
  await dialog.getByRole("button", { name: "Create account" }).click();
  await dialog.getByLabel("Email").fill(email);
  await dialog.getByRole("button", { name: "Continue", exact: true }).click();
  await dialog.getByLabel("Password").fill(password);
  await dialog.getByRole("button", { name: "Create account" }).click();
  await expect(dialog).toBeHidden();
}

async function activateSeller(page: Page) {
  await page.getByRole("link", { name: "Become a seller" }).click();
  await expect(page).toHaveURL(/\/seller$/);
}

test("seller publishing, public reporting, and staff moderation work across the full stack", async ({ browser }) => {
  const sellerContext = await browser.newContext();
  const seller = await sellerContext.newPage();
  const sellerEmail = uniqueEmail("seller");
  const title = `E2E Cebu condo ${Date.now()}`;

  await register(seller, sellerEmail);
  await activateSeller(seller);

  await seller.getByLabel("Property type").selectOption("Condo");
  await seller.getByLabel("Listing title").fill(title);
  await seller.getByLabel("City").fill("Cebu City");
  await seller.getByLabel("Barangay").fill("Lahug");
  await seller.getByLabel("Asking price (PHP)").fill("5600000");
  await seller.getByLabel("Floor or lot area (m²)").fill("84");
  await seller.getByLabel("Bedrooms").fill("2");
  await seller.getByLabel("Bathrooms").fill("2");
  await seller.getByLabel("Description").fill("A complete condominium listing created by the end-to-end browser suite.");
  await seller.getByLabel("Seller or agent name").fill("E2E Seller");
  await seller.getByLabel("Phone number").fill("+63 917 555 0188");
  await seller.getByLabel("Email address").fill(sellerEmail);
  await seller.getByRole("button", { name: "Save draft" }).click();
  await expect(seller.getByText("Your draft is saved to your Hacuba seller account.")).toBeVisible();

  await seller.locator("#listing-photos").setInputFiles([
    { name: "front.jpg", mimeType: "image/jpeg", buffer: Buffer.from("e2e-front-image") },
    { name: "living-room.jpg", mimeType: "image/jpeg", buffer: Buffer.from("e2e-living-room-image") },
  ]);
  await expect(seller.getByText("2 of 2 selected photos uploaded and registered.")).toBeVisible();

  await seller.getByRole("button", { name: "Publish listing" }).click();
  const publicLink = seller.getByRole("link", { name: "Open the public listing" });
  await expect(publicLink).toBeVisible();
  const publicPath = await publicLink.getAttribute("href");
  expect(publicPath).toMatch(/^\/listings\/[0-9a-f-]+$/);

  await seller.reload();
  await seller.getByRole("button", { name: "Account menu" }).click();
  await expect(seller.getByText(`Signed in as ${sellerEmail}`)).toBeVisible();

  const buyerContext = await browser.newContext();
  const buyer = await buyerContext.newPage();
  await buyer.goto("/");
  await buyer.getByRole("button", { name: /Where/ }).click();
  await buyer.getByRole("button", { name: "Cebu City, Philippines" }).click();
  await buyer.getByRole("button", { name: /Price/ }).click();
  await buyer.getByRole("button", { name: "₱3M – ₱6M" }).click();
  await buyer.getByRole("button", { name: /Property Type/ }).click();
  await buyer.getByRole("button", { name: "Condo" }).click();
  await buyer.getByRole("button", { name: "Search properties" }).click();
  await expect(buyer).toHaveURL(/city=Cebu\+City.*type=condo.*max_price=600000000/);
  await expect(buyer.locator(`a[href="${publicPath}"]`)).toBeVisible();

  await buyer.goto(publicPath!);
  await expect(buyer.getByRole("heading", { name: title })).toBeVisible();
  await expect(buyer.getByText("1 / 2")).toBeVisible();
  await buyer.getByRole("button", { name: "Next photo" }).click();
  await expect(buyer.getByText("2 / 2")).toBeVisible();
  await expect(buyer.getByRole("link", { name: "+63 917 555 0188" })).toBeVisible();

  await buyer.getByRole("link", { name: "Report this listing" }).click();
  await buyer.getByRole("radio", { name: /Suspicious payment request/ }).check();
  const reportDescription = "The poster asked for a deposit before arranging an in-person viewing.";
  await buyer.getByLabel("What happened?").fill(reportDescription);
  await buyer.getByRole("button", { name: "Send report" }).click();
  await expect(buyer.getByRole("heading", { name: "Thank you for flagging this" })).toBeVisible();

  const staffContext = await browser.newContext();
  const staff = await staffContext.newPage();
  await register(staff, "e2e-staff@hacuba.test");
  await staff.getByRole("button", { name: "Account menu" }).click();
  await staff.getByRole("menuitem", { name: "Support queue" }).click();
  const report = staff.locator("article").filter({ hasText: reportDescription });
  await expect(report).toBeVisible();
  await report.getByLabel("Reason").fill("The report identified a payment request before a viewing.");
  await report.getByRole("button", { name: "Hide listing" }).click();
  await expect(staff.getByText("Listing hidden. Mark the report triaged or resolved when ready.")).toBeVisible();

  await buyer.goto(publicPath!);
  await expect(buyer.getByRole("heading", { name: "This property is no longer available." })).toBeVisible();

  await report.getByRole("button", { name: "Restore listing" }).click();
  await expect(staff.getByText("Listing restored.")).toBeVisible();
  const restored = await buyer.goto(publicPath!);
  expect(restored?.status()).toBe(200);
  await expect(buyer.getByRole("heading", { name: title })).toBeVisible();

  await seller.getByRole("button", { name: "Account menu" }).click();
  await seller.getByRole("menuitem", { name: "Log out" }).click();
  await seller.goto("/seller");
  await expect(seller.getByRole("alert")).toContainText("Log in and activate seller access");

  await Promise.all([sellerContext.close(), buyerContext.close(), staffContext.close()]);
});

test("seller validation preserves a draft and exposes every missing publish requirement", async ({ browser }) => {
  const context = await browser.newContext();
  const page = await context.newPage();

  await register(page, uniqueEmail("validation"));
  await activateSeller(page);

  await page.getByLabel("Listing title").fill("bad");
  await page.getByRole("button", { name: "Save draft" }).click();
  await expect(page.getByRole("alert")).toContainText("must be 5 to 120 characters");

  await page.getByLabel("Listing title").fill("Draft that remains private");
  await page.getByRole("button", { name: "Save draft" }).click();
  await expect(page.getByText("Your draft is saved to your Hacuba seller account.")).toBeVisible();
  await page.getByRole("button", { name: "Publish listing" }).click();
  const validation = page.getByRole("alert");
  await expect(validation).toContainText("Finish these items before publishing");
  await expect(validation).toContainText("at least one image is required");
  await expect(validation).toContainText("missing");

  await page.getByRole("link", { name: "My listings" }).click();
  await expect(page.getByText("Drafts (1)")).toBeVisible();
  await expect(page.getByRole("heading", { name: "Draft that remains private" })).toBeVisible();
  await context.close();
});

test("anonymous and buyer accounts cannot enter seller or staff workspaces", async ({ browser, page }) => {
  await page.goto("/listings/00000000-0000-0000-0000-000000000000");
  await expect(page.getByRole("heading", { name: "This property is no longer available." })).toBeVisible();

  await page.goto("/seller");
  await expect(page.getByRole("alert")).toContainText("Log in and activate seller access");

  await page.goto("/support");
  await expect(page.getByRole("heading", { name: "Staff access required" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Support queue" })).toHaveCount(0);

  const buyerContext = await browser.newContext();
  const buyer = await buyerContext.newPage();
  await register(buyer, uniqueEmail("buyer"));
  await buyer.goto("/support");
  await expect(buyer.getByRole("heading", { name: "Staff access required" })).toBeVisible();
  await buyer.getByRole("button", { name: "Account menu" }).click();
  await expect(buyer.getByRole("menuitem", { name: "Support queue" })).toHaveCount(0);
  await buyerContext.close();
});
