-- just a scratchpad


-- Activate Node
INSERT INTO organizations (name) VALUES ('example') ON CONFLICT DO NOTHING;
INSERT INTO registration_tokens (id, name, organization_id) SELECT '89e07b4e-3943-4ee1-8f06-e63b65892289', 'example_token', organizations.id FROM organizations WHERE name='example' ON CONFLICT DO NOTHING;

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