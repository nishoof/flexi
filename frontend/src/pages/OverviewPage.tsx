import React from 'react';
import AddEntryModal from '../components/AddEntryModal';
import StatCard from '../components/StatCard';

function OverviewPage() {
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = React.useState(false);

  return (
    <>
      <div className="flex space-x-4">
        <StatCard title="Flexi Remaining" value={67.67} />
        <StatCard title="Flexi Remaining" value={67.67} />
        <StatCard title="Flexi Remaining" value={67.67} />
        <StatCard title="Flexi Remaining" value={67.67} />
      </div>

      <AddEntryModal
        isOpen={isAddEntryModalOpen}
        close={() => setIsAddEntryModalOpen(false)}
      />
      <button onClick={() => setIsAddEntryModalOpen(true)} >
        Open modal
      </button>
    </>
  );
}

export default OverviewPage;
