# Listing moderation

Staff can hide a published listing only from a Support report that contains a
linked Hacuba listing UUID. The Listings service records the staff decision,
reason, report UUID, and time in its own audit table. A hidden listing is
unavailable from public browse and detail endpoints, while its owner can see
the decision reason in the seller workspace.

Restoring a hidden listing is also staff-only and creates a second audit row.
It restores the listing to `published`; it does not expose the report text,
contact email, or reporter identity to the seller.

The Support and Listings audit records are deliberately separate because the
services own separate databases. Staff must use the linked report UUID for a
moderation action, so the two records remain traceable without sharing either
service's database credentials.
