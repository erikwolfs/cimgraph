SET SCHEMA 'cimgraph'

DROP SEQUENCE IF EXISTS object_id_seq
;

DROP SEQUENCE IF EXISTS predicate_id_seq
;

DROP SEQUENCE IF EXISTS subject_id_seq
;

DROP SEQUENCE IF EXISTS dataset_id_seq
;

DROP SEQUENCE IF EXISTS dsattribute_id_seq
;

/* Drop Tables */

DROP TABLE IF EXISTS object CASCADE
;

DROP TABLE IF EXISTS predicate CASCADE
;

DROP TABLE IF EXISTS resource CASCADE
;

DROP TABLE IF EXISTS subject CASCADE
;

DROP TABLE IF EXISTS dsattribute CASCADE
;

DROP TABLE IF EXISTS subject CASCADE
;

/* Create Tables */

CREATE TABLE dataset
(
	id BIGINT NOT NULL   DEFAULT NEXTVAL(('"dataset_id_seq"'::text)::regclass),
	tstype VARCHAR(100) NOT NULL,
	rdfid VARCHAR(50) NOT NULL
)
;

CREATE TABLE dsattribute
(
	id BIGINT NOT NULL   DEFAULT NEXTVAL(('"dsattribute_id_seq"'::text)::regclass),
	atname VARCHAR(50) NOT NULL,
	atvalue TEXT NULL,
	dataset_id BIGINT NOT NULL
)
;

CREATE TABLE object
(
	id BIGINT NOT NULL   DEFAULT NEXTVAL(('"object_id_seq"'::text)::regclass),
	predicate_id BIGINT NOT NULL,
	value_str TEXT NULL,
	value_flt DOUBLE PRECISION NULL,
	resource_uri VARCHAR(50) NULL,
	valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
	valid_to TIMESTAMP WITH TIME ZONE NOT NULL,
	created_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
	modified_time TIMESTAMP WITHOUT TIME ZONE NULL
)
;

CREATE TABLE predicate
(
	id BIGINT NOT NULL   DEFAULT NEXTVAL(('"predicate_id_seq"'::text)::regclass),
	subject_id BIGINT NOT NULL,
	pretype VARCHAR(100) NOT NULL,
	created_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
	modified_time TIMESTAMP WITHOUT TIME ZONE NULL
)
;

CREATE TABLE resource
(
	object_id BIGINT NOT NULL,
	subject_id BIGINT NOT NULL,
	created_time TIMESTAMP WITHOUT TIME ZONE NOT NULL
)
;

CREATE TABLE subject
(
	id BIGINT NOT NULL   DEFAULT NEXTVAL(('"subject_id_seq"'::text)::regclass),
	dataset_id bigint NOT NULL,
	subtype VARCHAR(100) NOT NULL,
	rdfid VARCHAR(50) NOT NULL,
	valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
	valid_to TIMESTAMP WITH TIME ZONE NOT NULL,
	created_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
	modified_time TIMESTAMP WITHOUT TIME ZONE NULL
)
;

/* Create Primary Keys, Indexes, Uniques, Checks */

ALTER TABLE dataset ADD CONSTRAINT "PK_dataset"
	PRIMARY KEY (id)
;

ALTER TABLE dsattribute ADD CONSTRAINT "PK_dsattribute"
	PRIMARY KEY (id)
;

CREATE INDEX "IXFK_dsattribute_dataset" ON dsattribute (dataset_id ASC)
;

ALTER TABLE object ADD CONSTRAINT "PK_object"
	PRIMARY KEY (id)
;

CREATE INDEX "IXFK_object_predicate" ON object (predicate_id ASC)
;

ALTER TABLE predicate ADD CONSTRAINT "PK_predicate"
	PRIMARY KEY (id)
;

CREATE INDEX "IXFK_predicate_subject" ON predicate (subject_id ASC)
;

CREATE INDEX "IDX_predicate_type" ON predicate (pretype ASC)
;

ALTER TABLE resource ADD CONSTRAINT "PK_resource"
	PRIMARY KEY (object_id,subject_id)
;

CREATE INDEX "IXFK_resource_object" ON resource (object_id ASC)
;

CREATE INDEX "IXFK_resource_subject" ON resource (subject_id ASC)
;


ALTER TABLE subject ADD CONSTRAINT "PK_subject"
	PRIMARY KEY (id)
;

CREATE INDEX "IDX_subject_type" ON subject (subtype ASC)
;

CREATE INDEX "IDX_rdfid" ON subject (rdfid ASC)
;

CREATE INDEX "IXFK_subject_dataset" ON subject (dataset_id ASC)
;

/* Create Foreign Key Constraints */

ALTER TABLE object ADD CONSTRAINT "FK_object_predicate"
	FOREIGN KEY (predicate_id) REFERENCES predicate (id) ON DELETE No Action ON UPDATE No Action
;

ALTER TABLE predicate ADD CONSTRAINT "FK_predicate_subject"
	FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE No Action ON UPDATE No Action
;

ALTER TABLE resource ADD CONSTRAINT "FK_resource_object"
	FOREIGN KEY (object_id) REFERENCES object (id) ON DELETE No Action ON UPDATE No Action
;

ALTER TABLE resource ADD CONSTRAINT "FK_resource_subject"
	FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE No Action ON UPDATE No Action
;

ALTER TABLE dsattribute ADD CONSTRAINT "FK_dsattribute_dataset"
	FOREIGN KEY (dataset_id) REFERENCES dataset (id) ON DELETE No Action ON UPDATE No Action
;

ALTER TABLE subject ADD CONSTRAINT "FK_subject_dataset"
	FOREIGN KEY (dataset_id) REFERENCES dataset (id) ON DELETE No Action ON UPDATE No Action
;

/* Create Table Comments, Sequences for Autonumber Columns */

CREATE SEQUENCE object_id_seq INCREMENT 1 START 1
;

CREATE SEQUENCE predicate_id_seq INCREMENT 1 START 1
;

CREATE SEQUENCE subject_id_seq INCREMENT 1 START 1

;
CREATE SEQUENCE dataset_id_seq INCREMENT 1 START 1
;

CREATE SEQUENCE dsattribute_id_seq INCREMENT 1 START 1

DROP PROCEDURE insert_subject(subtype VARCHAR(100),
								dataset_id BIGINT,
                                rdfid VARCHAR(50),
                                valid_from TIMESTAMP WITH TIME ZONE,
                                valid_to TIMESTAMP WITH TIME ZONE,
                                ret_id INOUT BIGINT);

CREATE OR REPLACE PROCEDURE insert_subject(subtype VARCHAR(100),
								dataset_id BIGINT,
                                rdfid VARCHAR(50),
                                valid_from TIMESTAMP WITH TIME ZONE,
                                valid_to TIMESTAMP WITH TIME ZONE,
                                ret_id INOUT BIGINT)
LANGUAGE plpgsql
AS $$
BEGIN
	INSERT INTO subject(subtype,
						dataset_id,
                        rdfid,
                        valid_from,
                        valid_to,
						created_time)
	VALUES ($1,
            $2,
            $3,
            $4,
			$5,
			CURRENT_TIMESTAMP)
	RETURNING subject.id INTO ret_id;
END;
$$;

DROP PROCEDURE insert_predicate(subject_id BIGINT,
                                pretype VARCHAR(100),
                                ret_id INOUT BIGINT)

CREATE OR REPLACE PROCEDURE insert_predicate(subject_id BIGINT,
								pretype VARCHAR(100),
                                ret_id INOUT BIGINT)
LANGUAGE plpgsql
AS $$
BEGIN
	INSERT INTO predicate(subject_id,
						pretype,
						created_time)
	VALUES ($1,
            $2,
			CURRENT_TIMESTAMP)
	RETURNING predicate.id INTO ret_id;
END;
$$;

DROP PROCEDURE insert_object(predicate_id BIGINT,
                                value_str TEXT,
								value_flt DOUBLE PRECISION,
								resource_uri VARCHAR(50),
								valid_from TIMESTAMP WITH TIME ZONE,
                                valid_to TIMESTAMP WITH TIME ZONE,
                                red_id INOUT BIGINT)

CREATE OR REPLACE PROCEDURE insert_object(predicate_id BIGINT,
                                		value_str TEXT,
										value_flt DOUBLE PRECISION,
										resource_uri VARCHAR(50),
										valid_from TIMESTAMP WITH TIME ZONE,
                                		valid_to TIMESTAMP WITH TIME ZONE,
                                		ret_id INOUT BIGINT)
LANGUAGE plpgsql
AS $$
BEGIN
	INSERT INTO object(predicate_id,
						value_str,
						value_flt,
						resource_uri,
						valid_from,
						valid_to,
						created_time)
	VALUES ($1,
            $2,
			$3,
			$4,
			$5,
			$6,
			CURRENT_TIMESTAMP)
	RETURNING object.id INTO ret_id;
END;
$$;

DROP PROCEDURE insert_resource(object_id BIGINT,
                                subject_id BIGINT)

CREATE OR REPLACE PROCEDURE insert_resource(object_id BIGINT,
											subject_id BIGINT)
LANGUAGE plpgsql
AS $$
BEGIN
	INSERT INTO resource(object_id,
						subject_id,
						created_time)
	VALUES ($1,
            $2,
			CURRENT_TIMESTAMP);
END;
$$;

CREATE OR REPLACE PROCEDURE insert_dataset(tstype VARCHAR(100),
											rdfid VARCHAR(50),
											ret_id INOUT BIGINT)
LANGUAGE plpgsql
AS $$
BEGIN
	INSERT INTO dataset(tstype,
						rdfid)
	VALUES ($1,
            $2)
	RETURNING dataset.id INTO ret_id;
END;
$$;

CREATE OR REPLACE PROCEDURE insert_dsattribute(atname VARCHAR(50),
											atvalue TEXT,
											dataset_id BIGINT)
LANGUAGE plpgsql
AS $$
BEGIN
	INSERT INTO dsattribute(atname,
						atvalue,
						dataset_id)
	VALUES ($1,
            $2,
			$3);
END;
$$;

CREATE OR REPLACE FUNCTION cimgraph.retrieve_abouts_without_link()
 RETURNS TABLE(object_id bigint,
 				recourse_uri character varying,
				valid_from object.valid_from%type ,
				valid_to object.valid_to%type)
 LANGUAGE sql
AS $function$
	SELECT object.id,
	object.resource_uri,
	object.valid_from,
	object.valid_to
	FROM object
	LEFT JOIN resource ON object.id = resource.object_id
	WHERE object.resource_uri <> ''
	AND resource.object_id IS NULL;
$function$

CREATE OR REPLACE FUNCTION cimgraph.retrieve_subject_by_rdfid(rdfabout VARCHAR(50),
																validfrom TIMESTAMP WITH TIME ZONE,
																validto TIMESTAMP WITH TIME ZONE)
 RETURNS TABLE(subject_id bigint,
 			   dataset_id bigint,
			   subtype VARCHAR(100),
			   valid_from TIMESTAMP WITH TIME ZONE,
			   valid_to TIMESTAMP WITH TIME ZONE)
 LANGUAGE sql
AS $function$
	SELECT subject.id,
	subject.dataset_id,
	subject.subtype,
	subject.valid_from,
	subject.valid_to
	FROM subject
	WHERE subject.rdfid = rdfabout
	AND subject.valid_from <= validfrom
	AND subject.valid_to >= validto;
$function$

CREATE OR REPLACE FUNCTION cimgraph.retrieve_subjects_by_type(stype VARCHAR(100),
																validfrom TIMESTAMP WITH TIME ZONE,
																validto TIMESTAMP WITH TIME ZONE)
RETURNS TABLE(subject_id bigint,
			   subtype VARCHAR(100),
			   rdfid VARCHAR(50))
LANGUAGE sql
AS $function$
	SELECT subject.id,
	subject.subtype,
	subject.rdfid
	FROM subject
	WHERE subject.subtype = stype
	AND subject.valid_from <= validfrom
	AND subject.valid_to >= validto;
$function$

CREATE OR REPLACE FUNCTION cimgraph.retrieve_predicates_by_subject(subjectid BIGINT,
																validfrom TIMESTAMP WITH TIME ZONE,
																validto TIMESTAMP WITH TIME ZONE)
RETURNS TABLE(predicate_id BIGINT,
			   pretype VARCHAR(100),
			   object_id BIGINT,
			   value_str TEXT,
			   value_flt DOUBLE PRECISION,
			   resource_uri VARCHAR(50))
LANGUAGE SQL
AS $FUNCTION$
	SELECT predicate.id,
			predicate.pretype,
			object.id,
			object.value_str,
			object.value_flt,
			object.resource_uri
	FROM predicate
	INNER JOIN object ON predicate.id = object.predicate_id
	WHERE predicate.subject_id = subjectid
	AND object.valid_from <= validfrom
	AND object.valid_to >= validto;
$FUNCTION$

CREATE OR REPLACE FUNCTION cimgraph.retrieve_childs_by_subject(subjectid BIGINT,
																validfrom TIMESTAMP WITH TIME ZONE,
																validto TIMESTAMP WITH TIME ZONE)
RETURNS TABLE(child_subid BIGINT,
				child_subtype VARCHAR(100),
				child_rdfid VARCHAR(50))
LANGUAGE SQL
AS $FUNCTION$
	SELECT subject.id,
			subject.subtype,
			subject.rdfid
	FROM resource
	INNER JOIN object ON resource.object_id = object.id
	INNER JOIN predicate ON object.predicate_id = predicate.id
	INNER JOIN subject ON predicate.subject_id = subject.id
	WHERE resource.subject_id = subjectid
	AND object.valid_from <= validfrom
	AND object.valid_to >= validto;
$FUNCTION$

