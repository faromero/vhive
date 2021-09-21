import os
import sys
import glob
import numpy as np
import pandas as pd
from typing import NamedTuple

class TraceQuery(NamedTuple):
  endpoint: str
  execution_time: int
  memory: int
  object_size: int
  invocations: np.ndarray

class TraceManager(object):
  def __init__(self, endpoints, tracedir, filter_days, sort_traces):
    self.endpoints = endpoints
    self.filter_days = filter_days
    self.sort_traces = sort_traces

    invocations_list, memory_list, duration_list = self.__parseTraces(tracedir)
    self.invocations_list = invocations_list
    self.memory_list = memory_list
    self.duration_list = duration_list

  def generateQueries(self, min_range, max_range):
    endpoint_map, func_app_map = self.__mapEndpoints()
    invoc_df = self.__mapInvocations(list(endpoint_map.keys()), min_range, max_range)
    query_list = self.__generateQueryInputs(endpoint_map, func_app_map, invoc_df)

    return query_list

  # Function to generate query inputs
  def __generateQueryInputs(self, endpoint_map, func_app_map, invoc_df):
    query_list = []

    for k,v in endpoint_map.items():
      function_name = k
      function_app = func_app_map[function_name]
      endpoint = v

      # TODO: only considering the latency and memory of the first available day.
      # Other option is to average/99% across all days
      df_exec = self.duration_list[0]
      df_mem = self.memory_list[0]
      execution_time = df_exec.loc[df_exec.HashFunction == function_name]['percentile_Average_99'].iloc[0]
      memory = df_mem.loc[df_mem.HashApp == function_app]['AverageAllocatedMb_pct99'].iloc[0]

      # Extract pertinent row from invoc_df
      invoc_row = invoc_df[(invoc_df.HashFunction == function_name)].drop(['HashApp', 'HashFunction'], axis=1)
      invoc_row_np = invoc_row.to_numpy()[0]

      # Create the next query info
      next_query_info = TraceQuery(endpoint=endpoint, execution_time=execution_time,
                                   memory=memory, object_size=0, invocations=invoc_row_np)
      query_list.append(next_query_info)

    return query_list

  # Function to map the invocations to a range
  def __mapInvocations(self, function_names, min_range, max_range):
    joined_df = None
    for df in self.invocations_list:
      # Extract relevant rows
      relevant_df = df[df.HashFunction.isin(function_names)]

      if joined_df is None:
        joined_df = relevant_df
      else:
        joined_df = joined_df.merge(relevant_df, how='left', on=['HashApp', 'HashFunction'])

    # Find the max and min of the DF
    max_val = joined_df.select_dtypes(include=[np.number]).max().max()
    min_val = joined_df.select_dtypes(include=[np.number]).min().min()

    # Remap all values to desired range
    input_range = max_val - min_val
    output_range = max_range - min_range

    def newRange(inp):
      # Ignore if dtype is not int64
      # This allows us to keep the HashApp and HashFunction columns intact
      if inp.dtype != np.int64:
        return inp

      return np.round((inp - min_val)*output_range / input_range + min_range)

    joined_df = joined_df.apply(newRange)

    return joined_df

  # Function to sort traces as indicated by user
  # By the time this function is called, sort_traces will have been validated
  def __sorttraces(self, df):
    if self.sort_traces == "mintime":
      return df.sort_values(by=['percentile_Average_99'], ascending=True)
    elif self.sort_traces == "maxtime":
      return df.sort_values(by=['percentile_Average_99'], ascending=False)
    elif self.sort_traces == "minmem":
      return df.sort_values(by=['AverageAllocatedMb_pct99'], ascending=True)
    elif self.sort_traces == "maxmem":
      return df.sort_values(by=['AverageAllocatedMb_pct99'], ascending=False)
    else: # "default"

  # Function to map functions to endpoints, and map functions to app hash
  # The latter is to speed up searches for memory
  def __mapEndpoints(self):
    endpoint_map = {}
    func_app_map = {}
    for i in self.duration_list:
      for _, row in i.iterrows():
        hash_app = row['HashApp']
        hash_function = row['HashFunction']
        if len(self.endpoints) > 0:
          next_endpoint = self.endpoints[0]
          self.endpoints.pop(0)
          endpoint_map[hash_function] = next_endpoint
          func_app_map[hash_function] = hash_app
        else:
          break
      if len(self.endpoints) == 0:
        break

    return endpoint_map, func_app_map

  # Function to parse traces and generate dataframes
  def __parseTraces(self, tracedir):
    # Helper function to check whether a trace should be parsed
    def shouldParse(trace_name):
      if not self.filter_days:
        return True
      for d in self.filter_days:
        curr_day = 'd%02d' % d
        if curr_day in trace_name:
          return True
      return False

    all_traces = glob.glob(os.path.join(tracedir, '*.csv'))

    invocations_list = []
    memory_list = []
    duration_list = [] 
    for at in all_traces:
      # Check whether we should parse the next trace
      if not shouldParse(at):
        continue

      print('Now parsing:', os.path.basename(at))

      # Read into DF
      df = pd.read_csv(at)

      if 'app_memory' in at:
        # Keep columns: HashApp, AverageAllocatedMb_pct99
        df = df[['HashApp','AverageAllocatedMb_pct99']]
        memory_list.append(df)
      elif 'function_durations' in at:
        # Keep columns: HashApp, HashFunction, percentile_Average_99
        df = df[['HashApp', 'HashFunction', 'percentile_Average_99']]
        duration_list.append(df)
      elif 'invocations_per_function' in at:
        # (Easier to drop.) Drop columns: HashOwner, Trigger
        df = df.drop(['HashOwner', 'Trigger'], axis=1)
        invocations_list.append(df)

    return invocations_list, memory_list, duration_list
