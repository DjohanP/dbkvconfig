package dbkvconfig

const getConfigByKey = `
	SELECT 
		COALESCE (%s, '') as value
	FROM 
		%s
	WHERE
		%s=$1
	LIMIT 1
`

const insertConfigQuery = `
	INSERT INTO
		%s
		(
			%s,
			%s
		)
	VALUES($1,$2)	
`

const updateConfigQuery = `
	UPDATE
		%s
	SET
		%s=$2
	WHERE
		%s=$1	

`
