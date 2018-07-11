package ladonauth

// PolicyStatements contains all policy related DB statements
type PolicyStatements struct {
	QueryInsertPolicy             string
	QueryInsertPolicyActions      string
	QueryInsertPolicyActionsRel   string
	QueryInsertPolicyResources    string
	QueryInsertPolicyResourcesRel string
	QueryInsertPolicySubjects     string
	QueryInsertPolicySubjectsRel  string
	QueryRequestCandidates        string
	// internal queries
	GetQuery     string
	GetAllQuery  string
	DeletePolicy string
}

// GetPolicyStatements returns statements specific to the db driver type
func GetPolicyStatements(driverName string) *PolicyStatements {
	stmts := &PolicyStatements{}
	switch driverName {
	case "postgres", "pg", "pgx":
		stmts.QueryInsertPolicy = `INSERT INTO am.ladon_policy(id, description, effect, conditions, meta) SELECT $1::varchar, $2, $3, $4, $5 WHERE NOT EXISTS (SELECT 1 FROM am.ladon_policy WHERE id = $1)`
		stmts.QueryInsertPolicyActions = `INSERT INTO am.ladon_action (id, template, compiled, has_regex) SELECT $1::varchar, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM am.ladon_action WHERE id = $1)`
		stmts.QueryInsertPolicyActionsRel = `INSERT INTO am.ladon_policy_action_rel (policy, action) SELECT $1::varchar, $2::varchar WHERE NOT EXISTS (SELECT 1 FROM am.ladon_policy_action_rel WHERE policy = $1 AND action = $2)`
		stmts.QueryInsertPolicyResources = `INSERT INTO am.ladon_resource (id, template, compiled, has_regex) SELECT $1::varchar, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM am.ladon_resource WHERE id = $1)`
		stmts.QueryInsertPolicyResourcesRel = `INSERT INTO am.ladon_policy_resource_rel (policy, resource) SELECT $1::varchar, $2::varchar WHERE NOT EXISTS (SELECT 1 FROM am.ladon_policy_resource_rel WHERE policy = $1 AND resource = $2)`
		stmts.QueryInsertPolicySubjects = `INSERT INTO am.ladon_subject (id, template, compiled, has_regex) SELECT $1::varchar, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM am.ladon_subject WHERE id = $1)`
		stmts.QueryInsertPolicySubjectsRel = `INSERT INTO am.ladon_policy_subject_rel (policy, subject) SELECT $1::varchar, $2::varchar WHERE NOT EXISTS (SELECT 1 FROM am.ladon_policy_subject_rel WHERE policy = $1 AND subject = $2)`
		stmts.QueryRequestCandidates = `
		SELECT
			p.id,
			p.effect,
			p.conditions,
			p.description,
			p.meta,
			subject.template AS subject,
			resource.template AS resource,
			action.template AS action
		FROM
			am.ladon_policy AS p

			INNER JOIN am.ladon_policy_subject_rel AS rs ON rs.policy = p.id
			LEFT JOIN am.ladon_policy_action_rel AS ra ON ra.policy = p.id
			LEFT JOIN am.ladon_policy_resource_rel AS rr ON rr.policy = p.id

			INNER JOIN am.ladon_subject AS subject ON rs.subject = subject.id
			LEFT JOIN am.ladon_action AS action ON ra.action = action.id
			LEFT JOIN am.ladon_resource AS resource ON rr.resource = resource.id
		WHERE
			(subject.has_regex IS NOT TRUE AND subject.template = $1)
			OR
			(subject.has_regex IS TRUE AND $2 ~ subject.compiled)`
		stmts.GetQuery = `SELECT
			p.id, p.effect, p.conditions, p.description, p.meta,
			subject.template as subject, resource.template as resource, action.template as action
		FROM
			am.ladon_policy as p
		
		LEFT JOIN am.ladon_policy_subject_rel as rs ON rs.policy = p.id
		LEFT JOIN am.ladon_policy_action_rel as ra ON ra.policy = p.id
		LEFT JOIN am.ladon_policy_resource_rel as rr ON rr.policy = p.id
		
		LEFT JOIN am.ladon_subject as subject ON rs.subject = subject.id
		LEFT JOIN am.ladon_action as action ON ra.action = action.id
		LEFT JOIN am.ladon_resource as resource ON rr.resource = resource.id
		
		WHERE p.id=$1`
		stmts.GetAllQuery = `SELECT
	p.id, p.effect, p.conditions, p.description, p.meta,
	subject.template as subject, resource.template as resource, action.template as action
FROM
	(SELECT * from am.ladon_policy ORDER BY id LIMIT $1 OFFSET $2) as p

LEFT JOIN am.ladon_policy_subject_rel as rs ON rs.policy = p.id
LEFT JOIN am.ladon_policy_action_rel as ra ON ra.policy = p.id
LEFT JOIN am.ladon_policy_resource_rel as rr ON rr.policy = p.id

LEFT JOIN am.ladon_subject as subject ON rs.subject = subject.id
LEFT JOIN am.ladon_action as action ON ra.action = action.id
LEFT JOIN am.ladon_resource as resource ON rr.resource = resource.id`
		stmts.DeletePolicy = "DELETE FROM am.ladon_policy WHERE id=$1"
	case "mysql":
		stmts.QueryInsertPolicy = `INSERT IGNORE INTO am.ladon_policy (id, description, effect, conditions, meta) VALUES(?,?,?,?,?)`
		stmts.QueryInsertPolicyActions = `INSERT IGNORE INTO am.ladon_action (id, template, compiled, has_regex) VALUES(?,?,?,?)`
		stmts.QueryInsertPolicyActionsRel = `INSERT IGNORE INTO am.ladon_policy_action_rel (policy, action) VALUES(?,?)`
		stmts.QueryInsertPolicyResources = `INSERT IGNORE INTO am.ladon_resource (id, template, compiled, has_regex) VALUES(?,?,?,?)`
		stmts.QueryInsertPolicyResourcesRel = `INSERT IGNORE INTO am.ladon_policy_resource_rel (policy, resource) VALUES(?,?)`
		stmts.QueryInsertPolicySubjects = `INSERT IGNORE INTO am.ladon_subject (id, template, compiled, has_regex) VALUES(?,?,?,?)`
		stmts.QueryInsertPolicySubjectsRel = `INSERT IGNORE INTO am.ladon_policy_subject_rel (policy, subject) VALUES(?,?)`
		stmts.QueryRequestCandidates = `
	SELECT
		p.id,
		p.effect,
		p.conditions,
		p.description,
		p.meta,
		subject.template AS subject,
		resource.template AS resource,
		action.template AS action
	FROM
		am.ladon_policy AS p

		INNER JOIN am.ladon_policy_subject_rel AS rs ON rs.policy = p.id
		LEFT JOIN am.ladon_policy_action_rel AS ra ON ra.policy = p.id
		LEFT JOIN am.ladon_policy_resource_rel AS rr ON rr.policy = p.id

		INNER JOIN am.ladon_subject AS subject ON rs.subject = subject.id
		LEFT JOIN am.ladon_action AS action ON ra.action = action.id
		LEFT JOIN am.ladon_resource AS resource ON rr.resource = resource.id
	WHERE
		(subject.has_regex = 0 AND subject.template = ?)
		OR
		(subject.has_regex = 1 AND ? REGEXP BINARY subject.compiled)`
		stmts.GetQuery = `SELECT
		p.id, p.effect, p.conditions, p.description, p.meta,
		subject.template as subject, resource.template as resource, action.template as action
	FROM
		am.ladon_policy as p
	
	LEFT JOIN am.ladon_policy_subject_rel as rs ON rs.policy = p.id
	LEFT JOIN am.ladon_policy_action_rel as ra ON ra.policy = p.id
	LEFT JOIN am.ladon_policy_resource_rel as rr ON rr.policy = p.id
	
	LEFT JOIN am.ladon_subject as subject ON rs.subject = subject.id
	LEFT JOIN am.ladon_action as action ON ra.action = action.id
	LEFT JOIN am.ladon_resource as resource ON rr.resource = resource.id
	
	WHERE p.id=?`
		stmts.GetAllQuery = `SELECT
	p.id, p.effect, p.conditions, p.description, p.meta,
	subject.template as subject, resource.template as resource, action.template as action
FROM
	(SELECT * from am.ladon_policy ORDER BY id LIMIT ? OFFSET ?) as p

LEFT JOIN am.ladon_policy_subject_rel as rs ON rs.policy = p.id
LEFT JOIN am.ladon_policy_action_rel as ra ON ra.policy = p.id
LEFT JOIN am.ladon_policy_resource_rel as rr ON rr.policy = p.id

LEFT JOIN am.ladon_subject as subject ON rs.subject = subject.id
LEFT JOIN am.ladon_action as action ON ra.action = action.id
LEFT JOIN am.ladon_resource as resource ON rr.resource = resource.id`
		stmts.DeletePolicy = "DELETE FROM am.ladon_policy WHERE id=?"
	default:
		return nil
	}
	return stmts
}
