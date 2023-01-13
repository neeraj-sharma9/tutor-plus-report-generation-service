package constant

const WEEKLY_REPORT = "WEEKLY_REPORT"
const WEEKLY_REPORT_JOB_TOPIC = "byjus-tutor-plus-wpr-job"
const WEEKLY_REPORT_PRIORITY_JOB_TOPIC = "byjus-tutor-plus-wpr-priority-job"
const WEEKLY_REPORT_RETRY_JOB_TOPIC = "byjus-tutor-plus-wpr-retry-job"

const MPR = "MPR"
const MPR_JOB_TOPIC = "byjus-tutor-plus-mpr-job"
const MPR_PRIORITY_JOB_TOPIC = "byjus-tutor-plus-mpr-priority-job"
const MPR_RETRY_JOB_TOPIC = "byjus-tutor-plus-mpr-retry-job"

const KAFKA_CONSUMER_GROUP = "mpr-worker-worker-group"

const TIME_LAYOUT = "2006-01-02 15:04:05"

const ASYNC_KAFKA_CONSUMERS = 6
const RETRY = 0

const (
	NOT_ACTIVE                      = "NOT_ACTIVE"                      // user not active
	SUBMITTED                       = "SUBMITTED"                       // job submitted to queue
	COMPLETED                       = "SUCCEEDED"                       // job competed
	JOB_RECEIVED                    = "JOB_RECEIVED"                    // job received ar worker
	JSON_GENERATION_COMPLETED       = "JSON_GENERATION_COMPLETED"       // json generation completed at worker
	INIT_PDF_GENERATION             = "INIT_PDF_GENERATION"             // pdf generation initiated at lambda
	PDF_GENERATION_COMPLETED        = "PDF_GENERATION_COMPLETED"        // pdf generation completed at frontend lambda
	PROCESS_TO_SF                   = "PROCESS_TO_SF"                   // SENDING PDF to Sales force
	PROCESS_TO_SF_FAILED            = "PROCESS_TO_SF_FAILED"            // SENDING PDF to Sales force Failed
	FAILED                          = "FAILED"                          // JOB FAILED
	JSON_GENERATION_FAILED          = "JSON_GENERATION_FAILED"          // if failed in worker
	PDF_GENERATION_FAILED           = "PDF_GENERATION_FAILED"           // if failed on lambda
	RETRY_SCHEDULED                 = "RETRY_SCHEDULED"                 // move to this state when pushing to scheduler
	RETRY_INITIATED                 = "RETRY_INITIATED"                 // move to this state one get call from scheduler
	TUTOR_PLUS_DATA_API_5XX_ERROR   = "TUTOR_PLUS_DATA_API_5XX_ERROR"   // api call to tutor plus for data failed with 5xx error
	PDF_GENERATION_REQUEST_RECEIVED = "PDF_GENERATION_REQUEST_RECEIVED" // pdf generation request received at frontend lambda
	ADD_START_AND_END_DATE          = "ADD_START_AND_END_DATE"          // pdf generation request received at frontend lambda
)

var REPORT_TOPICS = []string{MPR_JOB_TOPIC, MPR_PRIORITY_JOB_TOPIC, MPR_RETRY_JOB_TOPIC, WEEKLY_REPORT_JOB_TOPIC, WEEKLY_REPORT_PRIORITY_JOB_TOPIC, WEEKLY_REPORT_RETRY_JOB_TOPIC}
