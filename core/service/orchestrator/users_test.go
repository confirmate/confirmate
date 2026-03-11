package orchestrator

/*
 * Here, tests will be implemented. Currently, I just visualize a workflow via comments here to know which endpoints we need

1. User is created in KeyCloak (NOT in the Orchestrator)
	Thus, we consider that we have a compliance manager (CM) already and a regular User (U) with role "Control Owner"
2. CM invites U to the system (in the UI, users are just "added", i.e. imported by the KeyCloak)
	Q1: Should the user be then be added in Orchestrator? Or is it enough to have the user in KeyCloak?
	Q2: Who can add users:
		- Compliance Manager
		- Expert Compliance Manager
		- Chief Information Security Officer (CISO)
		- Admin
3. Grant U access to a TOE (or Audit Scope)
	Precondition: U don't have access TOE
	Q1: What are the consequences when U is only assigned a TOE but not to an Audit Scope
4. CM assigns U a Control C
	Precondition: U cannot change C's workflow
*/
