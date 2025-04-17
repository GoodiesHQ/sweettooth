-- just a scratchpad


-- Activate Node
INSERT INTO organizations (name) VALUES ('example') ON CONFLICT DO NOTHING;
INSERT INTO registration_tokens (id, name, organization_id) SELECT '46f265bb-f068-4f1f-90c2-90e879b2542d', 'example_token', organizations.id FROM organizations WHERE name='example' ON CONFLICT DO NOTHING;

UPDATE nodes SET organization_id=(SELECT id FROM organizations WHERE name='example') WHERE hostname='DLN6048';
UPDATE nodes SET approved=true WHERE hostname='dln6048';

INSERT INTO package_jobs(
    node_id,
    group_id,
    organization_id,
    action,
    name,
    version
) VALUES (
    (SELECT id FROM nodes WHERE hostname='dln6048'),
    NULL,
    (SELECT id FROM organizations WHERE name='example'),
    2,
    'bitwarden',
    NULL
);

SELECT * FROM nodes;