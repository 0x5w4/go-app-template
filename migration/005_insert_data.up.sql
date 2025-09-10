START TRANSACTION;

INSERT INTO
    `companies` (`id`, `name`, `icon`, `admin_id`)
VALUES
    (1, 'PCS Payment', 'icon_a.png', 1);

INSERT INTO
    `users` (`id`, `company_id`, `username`, `email`, `password`, `fullname`)
VALUES
    (1, 1, 'admin', 'admin@example.com', '$2a$12$mT2MbijTWSMvjOhneMEvXe/79/qLne5wmr42R8u0zTa8XnacQBCES', 'admin');

INSERT INTO
    `roles` (`id`, `code`, `name`, `super_admin`, `description`)
VALUES
    (1, 'ADMIN', 'Administrator', true, 'Full system access'),
    (2, 'STAFF', 'Staff', false, 'Limited operational access');

INSERT INTO
    `user_roles` (`user_id`, `role_id`)
VALUES
    (1, 1);

INSERT INTO
    `permissions` (`id`, `code`, `name`, `description`)
VALUES
    (1, 'ACQUIRER.READ', 'Acquirer Read', 'Permission to read acquirer'),
    (2, 'ACQUIRER.CREATE', 'Acquirer Create', 'Permission to create acquirer'),
    (3, 'ACQUIRER.UPDATE', 'Acquirer Update', 'Permission to update acquirer'),
    (4, 'ACQUIRER.DELETE', 'Acquirer Delete', 'Permission to delete acquirer'),
    (5, 'BIN.READ', 'Bin Read', 'Permission to read bin'),
    (6, 'BIN.CREATE', 'Bin Create', 'Permission to create bin'),
    (7, 'BIN.UPDATE', 'Bin Update', 'Permission to update bin'),
    (8, 'BIN.DELETE', 'Bin Delete', 'Permission to delete bin'),
    (9, 'CLIENT.READ', 'Client Read', 'Permission to read client'),
    (10, 'CLIENT.CREATE', 'Client Create', 'Permission to create client'),
    (11, 'CLIENT.UPDATE', 'Client Update', 'Permission to update client'),
    (12, 'CLIENT.DELETE', 'Client Delete', 'Permission to delete client'),
    (13, 'COMPANY.READ', 'Company Read', 'Permission to read company'),
    (14, 'COMPANY.CREATE', 'Company Create', 'Permission to create company'),
    (15, 'COMPANY.UPDATE', 'Company Update', 'Permission to update company'),
    (16, 'COMPANY.DELETE', 'Company Delete', 'Permission to delete company'),
    (17, 'ENDPOINT.READ', 'Endpoint Read', 'Permission to read endpoint'),
    (18, 'ENDPOINT.CREATE', 'Endpoint Create', 'Permission to create endpoint'),
    (19, 'ENDPOINT.UPDATE', 'Endpoint Update', 'Permission to update endpoint'),
    (20, 'ENDPOINT.DELETE', 'Endpoint Delete', 'Permission to delete endpoint'),
    (21, 'GROUP.READ', 'Group Read', 'Permission to read group'),
    (22, 'GROUP.CREATE', 'Group Create', 'Permission to create group'),
    (23, 'GROUP.UPDATE', 'Group Update', 'Permission to update group'),
    (24, 'GROUP.DELETE', 'Group Delete', 'Permission to delete group'),
    (25, 'ISSUER.READ', 'Issuer Read', 'Permission to read issuer'),
    (26, 'ISSUER.CREATE', 'Issuer Create', 'Permission to create issuer'),
    (27, 'ISSUER.UPDATE', 'Issuer Update', 'Permission to update issuer'),
    (28, 'ISSUER.DELETE', 'Issuer Delete', 'Permission to delete issuer'),
    (29, 'FEATURE.READ', 'Feature Read', 'Permission to read feature'),
    (30, 'FEATURE.CREATE', 'Feature Create', 'Permission to create feature'),
    (31, 'FEATURE.UPDATE', 'Feature Update', 'Permission to update feature'),
    (32, 'FEATURE.DELETE', 'Feature Delete', 'Permission to delete feature'),
    (33, 'MERCHANT.READ', 'Merchant Read', 'Permission to read merchant'),
    (34, 'MERCHANT.CREATE', 'Merchant Create', 'Permission to create merchant'),
    (35, 'MERCHANT.UPDATE', 'Merchant Update', 'Permission to update merchant'),
    (36, 'MERCHANT.DELETE', 'Merchant Delete', 'Permission to delete merchant'),
    (37, 'PERMISSION.READ', 'Permission Read', 'Permission to read permission'),
    (38, 'PERMISSION.CREATE', 'Permission Create', 'Permission to create permission'),
    (39, 'PERMISSION.UPDATE', 'Permission Update', 'Permission to update permission'),
    (40, 'PERMISSION.DELETE', 'Permission Delete', 'Permission to delete permission'),
    (41, 'PRINCIPLE.READ', 'Principle Read', 'Permission to read principle'),
    (42, 'PRINCIPLE.CREATE', 'Principle Create', 'Permission to create principle'),
    (43, 'PRINCIPLE.UPDATE', 'Principle Update', 'Permission to update principle'),
    (44, 'PRINCIPLE.DELETE', 'Principle Delete', 'Permission to delete principle'),
    (45, 'PRODUCT.READ', 'Product Read', 'Permission to read product'),
    (46, 'PRODUCT.CREATE', 'Product Create', 'Permission to create product'),
    (47, 'PRODUCT.UPDATE', 'Product Update', 'Permission to update product'),
    (48, 'PRODUCT.DELETE', 'Product Delete', 'Permission to delete product'),
    (49, 'PROFILE.READ', 'Profile Read', 'Permission to read profile'),
    (50, 'PROFILE.CREATE', 'Profile Create', 'Permission to create profile'),
    (51, 'PROFILE.UPDATE', 'Profile Update', 'Permission to update profile'),
    (52, 'PROFILE.DELETE', 'Profile Delete', 'Permission to delete profile'),
    (53, 'ROLE.READ', 'Role Read', 'Permission to read role'),
    (54, 'ROLE.CREATE', 'Role Create', 'Permission to create role'),
    (55, 'ROLE.UPDATE', 'Role Update', 'Permission to update role'),
    (56, 'ROLE.DELETE', 'Role Delete', 'Permission to delete role'),
    (57, 'STOCK.READ', 'Stock Read', 'Permission to read stock'),
    (58, 'STOCK.CREATE', 'Stock Create', 'Permission to create stock'),
    (59, 'STOCK.UPDATE', 'Stock Update', 'Permission to update stock'),
    (60, 'STOCK.DELETE', 'Stock Delete', 'Permission to delete stock'),
    (61, 'HELP_SERVICE.READ', 'Help Service Read', 'Permission to read help service'),
    (62, 'HELP_SERVICE.CREATE', 'Help Service Create', 'Permission to create help service'),
    (63, 'HELP_SERVICE.UPDATE', 'Help Service Update', 'Permission to update help service'),
    (64, 'HELP_SERVICE.DELETE', 'Help Service Delete', 'Permission to delete help service'),
    (65, 'TAG.READ', 'Tag Read', 'Permission to read tag'),
    (66, 'TAG.CREATE', 'Tag Create', 'Permission to create tag'),
    (67, 'TAG.UPDATE', 'Tag Update', 'Permission to update tag'),
    (68, 'TAG.DELETE', 'Tag Delete', 'Permission to delete tag'),
    (69, 'USER.READ', 'User Read', 'Permission to read user'),
    (70, 'USER.CREATE', 'User Create', 'Permission to create user'),
    (71, 'USER.UPDATE', 'User Update', 'Permission to update user'),
    (72, 'USER.DELETE', 'User Delete', 'Permission to delete user');

INSERT INTO
    `role_permissions` (`permission_id`, `role_id`)
VALUES
    (1, 1),
    (2, 1),
    (3, 1),
    (4, 1),
    (5, 1),
    (6, 1),
    (7, 1),
    (8, 1),
    (9, 1),
    (10, 1),
    (11, 1),
    (12, 1),
    (13, 1),
    (14, 1),
    (15, 1),
    (16, 1),
    (17, 1),
    (18, 1),
    (19, 1),
    (20, 1),
    (21, 1),
    (22, 1),
    (23, 1),
    (24, 1),
    (25, 1),
    (26, 1),
    (27, 1),
    (28, 1),
    (29, 1),
    (30, 1),
    (31, 1),
    (32, 1),
    (33, 1),
    (34, 1),
    (35, 1),
    (36, 1),
    (37, 1),
    (38, 1),
    (39, 1),
    (40, 1),
    (41, 1),
    (42, 1),
    (43, 1),
    (44, 1),
    (45, 1),
    (46, 1),
    (47, 1),
    (48, 1),
    (49, 1),
    (50, 1),
    (51, 1),
    (52, 1),
    (53, 1),
    (54, 1),
    (55, 1),
    (56, 1),
    (57, 1),
    (58, 1),
    (59, 1),
    (60, 1),
    (61, 1),
    (62, 1),
    (63, 1),
    (64, 1),
    (65, 1),
    (66, 1),
    (67, 1),
    (68, 1),
    (69, 1),
    (70, 1),
    (71, 1),
    (72, 1),
    (1, 2),
    (5, 2),
    (9, 2),
    (13, 2),
    (17, 2),
    (21, 2),
    (25, 2),
    (29, 2),
    (33, 2);

INSERT INTO 
    `support_features` (`id`, `key`, `code`, `name`, `is_active`) 
VALUES
    (1, 'edc_checking', 'HS34132', 'EDC Checking', 1),
    (2, 'helpdesk', 'HS67138', 'Request Paper Roll', 1),
    (3, 'call_hotline', 'HS40627', 'Kontak Helpdesk', 1),
    (5, 'inject_keksam', 'HS07361', 'Inject Keksam', 1),
    (6, 'about', 'HS27966', 'About', 1),
    (7, 'connection_test', 'HS16234', 'Connection Test', 1),
    (9, 'faq', 'HS03956', 'FAQ', 1),
    (10, 'lkm', 'HS68835', 'LKM', 1),
    (11, 'inbox', 'HS02469', 'Inbox', 1),
    (14, 'function', 'HS98807', 'Function', 1),
    (15, 'clear', 'HS94732', 'Clear', 1),
    (18, 'refresh', 'HS88309', 'Refresh', 1),
    (19, 'edc_store', 'HS96532', 'App Store', 1),
    (20, 'setting', 'HS98837', 'Setting', 1),
    (21, 'check_transaction', 'HS98838', 'Check Transaction', 1),
    (22, 'kerja_yuk', 'HS98839', 'KerjaYuk', 1),
    (23, 'pose', 'HS98840', 'POSe', 1);

INSERT INTO 
    `clients` (`id`, `company_id`, `code`, `name`, `phone`, `fax`, `icon`, `pic_name`, `pic_phone`, `district_id`, `village`, `postal_code`, `address`) 
VALUES
    (1, 1, 'CL076435', 'Bank Rakyat Indonesia (BRI)', '021-2510244', '021-2510255', 'https://placehold.co/600x400/00529B/FFFFFF?text=BRI', 'Budi Santoso', '081298765432', 317402, 'Tanah Abang', '10270', 'Gedung BRI I, Jl. Jend. Sudirman Kav.44-46'),
    (2, 1, 'CL345006', 'Bank Central Asia (BCA)', '021-23588000', '021-23588999', 'https://placehold.co/600x400/0033A0/FFFFFF?text=BCA', 'Sarah Anggraini', '081312345678', 317402, 'Menteng', '10310', 'Menara BCA, Jl. MH Thamrin No. 1'),
    (3, 1, 'CL123456', 'Bank Mandiri', '021-52997777', '021-52997778', 'https://placehold.co/600x400/00529B/FFFFFF?text=Mandiri', 'Rani Kusuma', '081212345678', 317402, 'Setiabudi', '12910', 'Plaza Mandiri, Jl. Jend. Gatot Subroto Kav.36-38');

INSERT INTO 
    `main_features` (`id`, `key`, `code`, `name`, `is_active`) 
VALUES
    (1, 'sale', 'FT671428', 'Sale', 1),
    (2, 'void', 'FT765942', 'Void', 1),
    (3, 'settlement', 'FT707169', 'Settlement', 1),
    (4, 'logon', 'FT914003', 'Logon', 1),
    (5, 'qris', 'FT460637', 'QRIS', 1),
    (6, 'update_balance', 'FT426707', 'Update Balance', 1),
    (7, 'card_info', 'FT531746', 'Card Info', 1),
    (8, 'payment', 'FT841586', 'Payment', 1),
    (9, 'online_topup', 'FT088051', 'Online Top Up', 1),
    (10, 'delayed_topup', 'FT385714', 'Delayed Top Up', 1),
    (11, 'balance_info', 'FT780567', 'Balance Info', 1),
    (12, 'delayed_balance', 'FT198258', 'Delayed Balance', 1),
    (13, 'initialize', 'FT114579', 'Initialize', 1),
    (14, 'print', 'FT541931', 'Print', 1),
    (15, 'print_log', 'FT698071', 'Print Log', 1),
    (16, 'report', 'FT143508', 'Report', 1),
    (17, 'credit', 'FT586431', 'Credit', 1),
    (18, 'sale_completion', 'FT755094', 'Sale Completion', 1),
    (19, 'card_verification', 'FT408083', 'Card Verification', 1),
    (20, 'sale_fare_non_fare', 'FT041615', 'Sale Fare Non Fare', 1),
    (21, 'sale_redemption', 'FT482620', 'Sale Redemption', 1),
    (22, 'manual_key_in', 'FT476278', 'Manual Key In', 1),
    (23, 'generate_qr', 'FT018473', 'Generate QR', 1),
    (24, 'status_transaksi_qr', 'FT886930', 'Status Transaksi', 1),
    (25, 'refund_qr', 'FT382540', 'Refund', 1),
    (26, 'helpdesk', 'FT010176', 'Helpdesk', 1),
    (27, 'call_hotline', 'FT779693', 'Call Hotline', 1),
    (28, 'random_pinpad', 'FT914004', 'Random Pinpad', 1),
    (29, 'sale_contactless', 'FT914005', 'Contactless', 1),
    (30, 'sale_tip', 'FT914006', 'Sale Tip', 1),
    (31, 'batch_upload', 'FT914007', 'Batch Upload', 1),
    (32, 'tip_qris', 'FT914008', 'Tip QRIS', 1),
    (33, 'nfc_payment', 'FT914009', 'QRIS Tap', 1),
    (34, 'nfc_refund', 'FT914010', 'Refund QRIS Tap', 1),
    (35, 'sale_fuel', 'FT914011', 'Sale Fuel Card', 1),
    (36, 'sale_point', 'FT914012', 'Sale Point', 1),
    (37, 'card_release_verif', 'FT914013', 'Card Release Verification', 1);

INSERT INTO
    `client_main_features` (`client_id`, `main_feature_id`, `tag_id`, `order`)
VALUES
    -- BRI Client Features
    (1, 1, 1, 1),  -- Sale
    (1, 2, 5, 2),  -- Void
    (1, 5, 1, 3),  -- QRIS Payment
    (1, 6, 2, 4),  -- Update Balance
    -- BCA Client Features
    (2, 17, 1, 1),  -- Credit
    (2, 18, 1, 2),  -- Sale Completion
    (2, 19, 3, 3),  -- Card Verification
    (2, 20, 4, 4),  -- Sale Fare Non Fare
    -- MANDIRI Client Features
    (3, 33, 1, 1),  -- NFC Payment
    (3, 34, 2, 2),  -- NFC Refund
    (3, 35, 3, 3),  -- Sale Fuel Card
    (3, 36, 4, 4);  -- Sale Point

INSERT INTO
    `client_support_features` (`client_id`, `support_feature_id`, `order`)
VALUES
    -- Assigning some support features to BRI Client
    (1, 1, 1),
    (1, 2, 2),
    (1, 3, 3),
    -- Assigning some support features to BCA Client
    (2, 1, 1),
    (2, 6, 2),
    (2, 7, 3),
    -- Assigning some support features to MANDIRI Client
    (3, 1, 1),
    (3, 2, 2),
    (3, 3, 3);

COMMIT;