import { JobsView } from "@/components/features/jobs-view";
import { apiFetch } from "@/lib/api-client";
import {
  jobRepoNameMap,
  jobRepoProviderMap,
  mapJob,
  unwrapApiData,
  type ApiJob,
  type ApiRepository,
} from "@/lib/map-api";

export async function JobsData() {
  const [jobsRes, reposRes] = await Promise.all([
    apiFetch("/jobs"),
    apiFetch("/repositories"),
  ]);

  const jobsParsed = jobsRes.ok
    ? unwrapApiData<ApiJob[]>(await jobsRes.json())
    : [];
  const reposParsed = reposRes.ok
    ? unwrapApiData<ApiRepository[]>(await reposRes.json())
    : [];
  const jobsRaw: ApiJob[] = Array.isArray(jobsParsed) ? jobsParsed : [];
  const reposRaw: ApiRepository[] = Array.isArray(reposParsed)
    ? reposParsed
    : [];
  const names = jobRepoNameMap(reposRaw);
  const providers = jobRepoProviderMap(reposRaw);
  const jobs = jobsRaw.map((j) =>
    mapJob(j, names[j.repository_id], providers[j.repository_id])
  );

  return <JobsView jobs={jobs} />;
}
